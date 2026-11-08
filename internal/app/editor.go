package app

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

type sheetState struct {
	create        bool
	scope         Scope
	originalName  string
	originalValue string
	name          string
	value         string
	valueMode     int // 0 text, 1 list
	entries       []string
	nameError     string
	returnFocus   ui.Handle
}

type batchState struct {
	scope         Scope
	target        string
	originalValue string
	text          string
	returnFocus   ui.Handle
}

// List editing preserves empty items and whitespace. Display/search parsing
// deliberately lives separately: opening an editor must never rewrite a value.
func (s *sheetState) currentValue() string {
	if s.valueMode == 1 {
		return strings.Join(s.entries, ";")
	}
	return s.value
}

func (s *sheetState) syncValueMode() {
	if s.valueMode == 1 {
		s.entries = strings.Split(s.value, ";")
	} else {
		s.value = strings.Join(s.entries, ";")
	}
}

func (a *App) sheetDirty() bool {
	s := a.sheet
	if s == nil {
		return false
	}
	if s.create {
		return strings.TrimSpace(s.name) != "" || s.currentValue() != ""
	}
	return strings.TrimSpace(s.name) != s.originalName || s.currentValue() != s.originalValue
}

func (a *App) batchDirty() bool {
	return a.batch != nil && analyzeBatch(a.batch.originalValue, a.batch.text).hasChanges
}

// confirmDiscard is shared by editor navigation and native window closing.
// The continuation runs only after approval, or immediately for a clean editor.
func (a *App) confirmDiscard(dirty bool, title, message, confirm string, next func()) {
	if !dirty {
		next()
		return
	}
	a.modal = &modalState{
		icon: iconInfo, title: title, message: message,
		confirm: confirm, cancel: "继续编辑", danger: true, onConfirm: next,
	}
	if a.win != nil {
		a.win.Invalidate()
	}
}

func (a *App) closeBatch() {
	if a.batch == nil {
		return
	}
	a.confirmDiscard(a.batchDirty(), "放弃批量编辑？", "粘贴的内容还没有应用，离开后会丢失。", "放弃更改", func() { a.batch = nil })
}

func (a *App) switchBatchTarget(name string) {
	b := a.batch
	if b == nil || name == b.target {
		return
	}
	target, ok := a.findVariable(b.scope, name)
	if !ok {
		return
	}
	a.confirmDiscard(a.batchDirty(), "切换变量？", "切换后，当前粘贴的内容不会保留。", "放弃并切换", func() {
		b.target = target.Name
		b.originalValue = target.Value
		b.text = strings.Join(splitValue(target.Value), "\n")
		if target.Value == "" {
			b.text = ""
		}
	})
}

// allowClose is also used for application quit, which emits window close events.
func (a *App) allowClose(closeWindow func()) bool {
	if !a.sheetDirty() && !a.batchDirty() {
		return true
	}
	a.confirmDiscard(true, "放弃更改并退出？", "尚有未保存的编辑，退出后会丢失。", "放弃并退出", func() {
		a.sheet = nil
		a.batch = nil
		closeWindow()
	})
	return false
}
