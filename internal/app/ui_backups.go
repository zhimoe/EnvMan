package app

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

func (a *App) backupsView(c *ui.Context) {
	if !a.backupsOpen {
		return
	}
	p := a.pal
	ui.Overlay(c, func() {
		back := ui.Row(c).Absolute().Left(0).Top(a.metrics.titleBarH).Right(0).Bottom(0).Center().
			Background(ui.RGBA(0, 0, 0, 0.38)).Modal()
		back.Children(func() {
			panel := ui.Column(c).Width(500).Padding(space16).Radius(radiusOverlay).Background(p.surface).
				Border(1, p.line).Shadow(0, 18, 50, 0, ui.RGBA(0, 0, 0, 0.35)).
				Role(ui.RoleDialog).Label("备份历史")
			panel.Children(func() {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(space8).Children(func() {
					ui.Row(c).Size(32, 32).Radius(tileRadius(32)).Background(p.accentSoft).Center().Children(func() {
						ui.Icon(c, iconClock).Size(16, 16).TextColor(p.accent)
					})
					ui.Column(c).Grow(1).Gap(space2).Children(func() {
						ui.Text(c, "备份历史").FontSize(a.metrics.fontSubtitle).FontWeight(weightSemibold).TextColor(p.text)
						ui.Text(c, "每次应用更改前自动创建快照，最多保留 20 份。").FontSize(a.metrics.fontCaption).TextColor(p.text3)
					})
					a.iconButton(c, iconX, "关闭", a.metrics.controlXs, func() { a.backupsOpen = false })
				})
				ui.Box(c).Height(space12)
				if a.snapshotsError != "" {
					ui.Text(c, a.snapshotsError).TextColor(p.danger)
				} else if len(a.snapshots) == 0 {
					ui.Column(c).FillWidth().Padding(space32).Center().Gap(space8).Children(func() {
						ui.Icon(c, iconClock).Size(22, 22).TextColor(p.text3)
						ui.Text(c, "还没有备份").FontSize(a.metrics.fontBody).FontWeight(weightSemibold).TextColor(p.text2)
						ui.Text(c, "应用更改后，这里会出现快照。").FontSize(a.metrics.fontSmall).TextColor(p.text3)
					})
				} else {
					ui.Scroll(c).MaxHeight(430).Children(func() {
						ui.Column(c).Gap(space4).Children(func() {
							for _, snap := range a.snapshots {
								a.backupRow(c, snap)
							}
						})
					})
				}
				ui.Box(c).Height(space12)
				ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
					a.button(c, "关闭", nil, buttonSecondary, true, func() { a.backupsOpen = false })
				})
			})
			if panel.PressedOutside() || back.OverlayShortcut(0, ui.KeyEscape) {
				a.backupsOpen = false
			}
		})
	})
}

func (a *App) backupRow(c *ui.Context, snap Snapshot) {
	p := a.pal
	ui.Row(c).FillWidth().Padding(space8, space8).Radius(radiusOverlay).Background(p.surface2).Border(1, p.line).
		AlignItems(ui.Center).Gap(space8).Children(func() {
		ui.Row(c).Size(27, 27).Radius(tileRadius(27)).Background(p.surface).Border(1, p.line).Center().Children(func() {
			ui.Icon(c, iconClock).Size(13, 13).TextColor(p.text3)
		})
		ui.Column(c).Grow(1).Gap(space2).Children(func() {
			ui.Row(c).Gap(space6).AlignItems(ui.Center).Children(func() {
				note := snap.Note
				if note == "" {
					note = "环境变量快照"
				}
				ui.Text(c, note).FontSize(a.metrics.fontSmall).FontWeight(weightSemibold).TextColor(p.text).SingleLine()
				a.scopePill(c, snap.Scope)
			})
			ui.Textf(c, "%s · %d 个变量", snap.Time.Local().Format("2006-01-02 15:04"), len(snap.Variables)).
				FontSize(a.metrics.fontCaption).TextColor(p.text3)
		})
		a.textButton(c, "恢复", true, func() {
			a.confirmRestore(c, snap)
		})
	})
}

func (a *App) confirmRestore(c *ui.Context, snap Snapshot) {
	t := snap.Time.Local().Format("2006-01-02 15:04")
	names := make([]string, 0, len(snap.Variables))
	for _, v := range snap.Variables {
		names = append(names, v.Name)
	}
	focus := ""
	if len(names) > 0 {
		focus = names[0]
		for _, n := range names {
			if strings.EqualFold(n, "Path") {
				focus = n
				break
			}
		}
	}
	a.modal = &modalState{
		icon:    iconClock,
		title:   "恢复此快照？",
		message: fmt.Sprintf("将把%s恢复到 %s 的状态，共 %d 个变量。\n\n当前状态会先自动备份。", snap.Scope.String(), t, len(snap.Variables)),
		confirm: "恢复快照",
		cancel:  "取消",
		danger:  true,
		onConfirm: func() {
			before := cloneVars(a.variables(snap.Scope))
			a.backupsOpen = false
			a.commit(snap.Scope, cloneVars(snap.Variables), "恢复快照", "已恢复备份快照", focus, before)
		},
	}
}
