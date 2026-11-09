package app

import (
	"fmt"
	"os"
	"strings"

	"envman/internal/application"
	"envman/internal/environment"
	"envman/internal/storage"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type App struct {
	win     *mygo.Window
	env     environment.Backend
	store   *storage.Store
	service *application.Service

	scope    Scope
	filter   ViewFilter
	query    string
	sortDesc bool

	vars         map[Scope][]Variable
	loadErrors   map[Scope]string
	missingPaths map[string]bool
	admin        bool

	expanded map[string]bool
	checked  map[string]bool

	focusUser   int
	focusSystem int

	userList   ui.ListState
	systemList ui.ListState

	themeMode ThemeMode
	lastDark  bool
	pal       palette
	metrics   metrics

	search ui.Handle

	sheet          *sheetState
	batch          *batchState
	modal          *modalState
	backupsOpen    bool
	snapshots      []Snapshot
	snapshotsError string

	startupToast string
	pendingToast string
	pendingErr   bool
	toastAction  *toastRequest
}

type toastRequest struct {
	message string
	label   string
	action  func()
}

type modalState struct {
	title     string
	message   string
	confirm   string
	cancel    string
	danger    bool
	icon      *ui.SVG
	onConfirm func()
	onCancel  func()
}

func newAppWith(backend environment.Backend, store *storage.Store) *App {
	a := &App{
		env:         backend,
		store:       store,
		vars:        make(map[Scope][]Variable),
		loadErrors:  make(map[Scope]string),
		expanded:    make(map[string]bool),
		checked:     make(map[string]bool),
		focusUser:   -1,
		focusSystem: -1,
	}
	a.service = application.New(backend, store)
	settings := a.loadSettings()
	a.themeMode = parseThemeMode(settings.Theme)
	if t := os.Getenv("ENVMAN_THEME"); t != "" {
		a.themeMode = parseThemeMode(t)
	}
	a.admin = a.env.Elevated()
	if v := os.Getenv("ENVMAN_ADMIN"); v == "0" {
		a.admin = false
	} else if v == "1" {
		a.admin = true
	}

	var applied *PendingEdits

	p, pendingErr := a.store.ReadPending()
	if pendingErr != nil {
		a.startupToast = "读取待应用更改失败：" + pendingErr.Error()
		a.pendingErr = true
	} else if p != nil && a.env.Elevated() {
		// Consume the command before writing, so failures cannot silently replay on restart.
		if claimed, err := a.store.ClaimPending(); err != nil {
			a.startupToast = "无法领取待应用更改，未写入：" + err.Error()
			a.pendingErr = true
		} else {
			result, err := a.service.Apply(application.Change{Scope: p.Scope, Before: p.Before, Desired: p.Variables, Note: "提权后应用更改前备份"})
			if err != nil {
				if result.Snapshot.ID != "" {
					a.handleApplyFailure(p.Scope, result.Snapshot, err)
				} else {
					a.startupToast = err.Error()
					a.pendingErr = true
				}
			} else {
				if err := a.store.CompletePending(claimed); err != nil {
					a.toast("更改已应用，但无法清理恢复文件：" + err.Error())
				}
				applied = p
			}
		}
	}

	a.reload()
	a.expand(ScopeUser, "Path")
	if applied != nil {
		a.scope = applied.Scope
		if applied.Note != "" {
			a.startupToast = applied.Note
		} else {
			a.startupToast = "已应用系统变量更改"
		}
		if applied.Focus != "" {
			a.expand(applied.Scope, applied.Focus)
			rows := a.filtered(applied.Scope)
			for i, item := range rows {
				if strings.EqualFold(item.Name, applied.Focus) {
					a.setFocus(i)
					break
				}
			}
		}
	}
	return a
}

func (a *App) reload() {
	for _, s := range []Scope{ScopeUser, ScopeSystem} {
		vars, err := a.env.Load(s)
		if err != nil {
			a.loadErrors[s] = environment.DescribeError(s, err)
			continue
		}
		a.vars[s] = vars
		delete(a.loadErrors, s)
	}
	a.scanMissingPaths(statPath)
	snapshots, err := a.store.LoadSnapshots()
	if err != nil {
		a.snapshotsError = err.Error()
	} else {
		a.snapshots = snapshots
		a.snapshotsError = ""
	}
}

func (a *App) view(c *ui.Context) {
	t, p := a.buildTheme(c)
	c.SetTheme(t)
	a.pal = p
	a.lastDark = t.Dark
	a.shortcuts(c)

	ui.Column(c).Fill().Background(p.window).Children(func() {
		a.titlebar(c)
		if a.batch != nil {
			a.batchView(c)
		} else {
			a.mainView(c)
		}
	})

	if a.sheet != nil {
		a.sheetView(c)
	}
	if a.backupsOpen {
		a.backupsView(c)
	}
	if a.modal != nil {
		a.modalView(c)
	}

	if a.startupToast != "" {
		msg := a.startupToast
		a.startupToast = ""
		if a.pendingErr {
			c.AddToast(ui.Toast{Title: msg, Type: "error", Timeout: 6_000_000_000})
			a.pendingErr = false
		} else {
			c.Toast(msg)
		}
	}
	if a.pendingToast != "" {
		msg := a.pendingToast
		a.pendingToast = ""
		c.Toast(msg)
	}
	if a.toastAction != nil {
		t := a.toastAction
		a.toastAction = nil
		c.ToastAction(t.message, t.label, t.action)
	}
}

func (a *App) toast(msg string) {
	a.pendingToast = msg
	if a.win != nil {
		a.win.Invalidate()
	}
}

func (a *App) undoToast(msg string, action func()) {
	a.toastAction = &toastRequest{message: msg, label: "撤销", action: action}
	if a.win != nil {
		a.win.Invalidate()
	}
}

func (a *App) shortcuts(c *ui.Context) {
	c.OnShortcut(ui.Cmd, ui.KeyK, func() {
		a.search.RequestFocus(c)
	})
	c.OnShortcut(ui.Cmd, ui.KeyN, func() {
		if a.sheet == nil && a.batch == nil && a.modal == nil {
			a.openCreate("")
		}
	})
	c.OnShortcut(ui.Cmd, ui.KeyE, func() {
		if a.sheet == nil && a.batch == nil && a.modal == nil {
			a.openBatch("")
		}
	})
	// The sheet and batch pages handle Escape themselves.
	c.OnShortcut(0, ui.KeySlash, func() {
		if a.sheet == nil && a.batch == nil && a.modal == nil && a.backupsOpen == false {
			a.search.RequestFocus(c)
		}
	})
}

func (a *App) switchScope(s Scope) {
	if a.scope == s {
		return
	}
	a.scope = s
	a.clearChecked()
	a.setFocus(-1)
	a.sheet = nil
	a.batch = nil
}

func (a *App) switchView(v ViewFilter) {
	if a.filter == v {
		return
	}
	a.filter = v
	a.clearChecked()
	a.setFocus(-1)
	if v == ViewPath {
		for _, item := range a.variables(a.scope) {
			if strings.EqualFold(item.Name, "Path") {
				a.expand(a.scope, item.Name)
				break
			}
		}
	}
}

func (a *App) scopeSubtitle() string {
	if a.scope == ScopeUser {
		return "仅影响当前用户 · 新终端生效"
	}
	if a.admin {
		return "影响所有用户 · 管理员模式"
	}
	return "影响所有用户 · 需要管理员权限才能修改"
}

func (a *App) showKey() string {
	return fmt.Sprintf("%d\x00", a.scope)
}
