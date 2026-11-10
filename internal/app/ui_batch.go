package app

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// The measurements of the batch editor's text pane, besides the metrics the
// theme gives: they are shared by the layout and by the width and height
// computation below, so that the pane fills the viewport exactly when its
// longest line fits.
const (
	batchGutterGap    float32 = space8
	batchBoxPadV      float32 = space6
	batchBoxPadH      float32 = space2
	batchBoxBorder    float32 = 1
	batchBodyPadV     float32 = space12
	batchBodyPadH     float32 = space12
	batchDividerWidth float32 = 1
	batchLeftFlex     float32 = 1.12
	batchRightFlex    float32 = 0.88
)

func (a *App) openBatch(preferred string) {
	if a.batch != nil {
		return
	}
	targets := a.batchTargets(a.scope)
	if len(targets) == 0 {
		a.toast("当前范围没有可批量编辑的多值变量")
		return
	}
	target := targets[0]
	for _, t := range targets {
		if preferred != "" && strings.EqualFold(t.Name, preferred) {
			target = t
			break
		}
	}
	if preferred == "" {
		for _, t := range targets {
			if strings.EqualFold(t.Name, "Path") {
				target = t
				break
			}
		}
	}
	entries := splitValue(target.Value)
	text := strings.Join(entries, "\n")
	if target.Value == "" {
		text = ""
	}
	a.batch = &batchState{
		scope:         a.scope,
		target:        target.Name,
		originalValue: target.Value,
		text:          text,
	}
	a.sheet = nil
	a.backupsOpen = false
}

func (a *App) batchView(c *ui.Context) {
	b := a.batch
	if b == nil {
		return
	}
	p := a.pal
	info := analyzeBatch(b.originalValue, b.text)
	entries, ignored := batchParse(b.text)
	targets := a.batchTargets(b.scope)
	names := make([]string, 0, len(targets))
	for _, t := range targets {
		names = append(names, t.Name)
	}
	validTarget := false
	for _, n := range names {
		if n == b.target {
			validTarget = true
			break
		}
	}
	if !validTarget && len(names) > 0 {
		a.switchBatchTarget(names[0])
		b = a.batch
		info = analyzeBatch(b.originalValue, b.text)
		entries, ignored = batchParse(b.text)
	}

	ui.Column(c).Grow(1).FillWidth().FillHeight().Background(p.window).Children(func() {
		// Header
		ui.Row(c).Height(a.metrics.pageHeaderH).FillWidth().Padding(0, space12).Background(p.sidebar).
			BorderWidth(0, 0, 1, 0).BorderColor(p.line).AlignItems(ui.Center).Gap(space8).Children(func() {
			a.iconButton(c, iconArrowLeft, "返回变量列表（Esc）", a.metrics.controlXs, func() { a.closeBatch() })
			ui.Text(c, "批量编辑值").FontSize(a.metrics.fontSubtitle).FontWeight(weightSemibold).TextColor(p.text)
			a.scopePill(c, b.scope)
			ui.Box(c).Width(4)
			ui.Text(c, "变量").FontSize(a.metrics.fontCaption).TextColor(p.text3)
			oldTarget := b.target
			sel := ui.Select(c, &b.target, names).Width(170).Label("目标变量")
			if sel.Changed() && b.target != oldTarget {
				next := b.target
				b.target = oldTarget
				a.switchBatchTarget(next)
			}
			w, _ := c.Size()
			if w > 1020 {
				ui.Row(c).MarginStart(5).Gap(space4).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconInfo).Size(11, 11).TextColor(p.text3)
					ui.Text(c, "每行一个条目 · 支持 ; 与换行分隔").FontSize(a.metrics.fontCaption).TextColor(p.text3).SingleLine()
				})
			}
			ui.Spacer(c).Grow(1)
			a.button(c, "取消", nil, buttonSecondary, true, func() { a.closeBatch() })
			label := "应用更改"
			if info.hasChanges && len(entries) == 0 && len(info.original) > 0 {
				label = "清空变量"
			}
			a.button(c, label, nil, buttonPrimary, info.hasChanges, func() { a.applyBatch(c) })
		})

		// Body
		ui.Row(c).Grow(1).FillWidth().AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Grow(1.12).Basis(0).FillHeight().Background(p.surface).Children(func() {
				ui.Row(c).Height(a.metrics.paneHeaderH).FillWidth().Padding(0, space12).BorderWidth(0, 0, 1, 0).BorderColor(p.lineSoft).
					AlignItems(ui.Center).Gap(space8).Children(func() {
					ui.Textf(c, "当前 %d 个条目", len(splitValue(b.originalValue))).FontSize(a.metrics.fontCaption).TextColor(p.text3)
					ui.Spacer(c).Grow(1)
					a.quietButton(c, "整理为多行", iconList, true, func() {
						b.text = strings.Join(entries, "\n")
					})
				})
				ui.Column(c).Grow(1).FillWidth().Padding(batchBodyPadV, batchBodyPadH).Children(func() {
					lineCount := strings.Count(b.text, "\n") + 1
					textWidth, textHeight := a.batchEditorViewport(c, b.text)
					textWidth = a.batchEditorTextWidth(c, b.text)
					ui.Box(c).Grow(1).FillWidth().Background(p.surface).Border(1, p.line).Radius(radiusOverlay).Clip().Children(func() {
						ui.Scroll(c).Grow(1).FillWidth().Children(func() {
							ui.Row(c).FillWidth().AlignItems(ui.Start).
								Gap(batchGutterGap).Padding(batchBoxPadV, batchBoxPadH).Children(func() {
								ui.Column(c).Width(numberColumnWidth(c, lineCount, a.metrics.fontBody)).Gap(0).Children(func() {
									for i := 1; i <= lineCount; i++ {
										ui.Textf(c, "%d", i).Font(monoFont).FontSize(a.metrics.fontBody).
											FixedLineHeight(a.metrics.lineBody).Height(a.metrics.lineBody).
											Width(numberColumnWidth(c, lineCount, a.metrics.fontBody)).TextColor(p.text3).TextAlign(ui.End)
									}
								})
								ui.ScrollHorizontal(c).Grow(1).Children(func() {
									ui.TextAreaBase(c, &b.text).Width(textWidth).MinHeight(textHeight).
										Font(monoFont).FontSize(a.metrics.fontBody).
										FixedLineHeight(a.metrics.lineBody).NoWrap().
										Placeholder("C:\\Program Files\\Tool\\bin\nC:\\Users\\Ada\\.local\\bin\nC:\\Windows\\System32").
										Label("编辑变量值").AutoFocus()
								})
							})
						})
					})
				})
				ui.Row(c).Height(a.metrics.barH).FillWidth().Padding(0, space12).BorderWidth(1, 0, 0, 0).BorderColor(p.lineSoft).
					AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconShieldDone).Size(11, 11).TextColor(p.success)
					ui.Text(c, "应用前自动创建备份").FontSize(a.metrics.fontCaption).TextColor(p.text3).Grow(1)
					meta := fmt.Sprintf("%d 个条目", len(entries))
					if ignored > 0 {
						meta += fmt.Sprintf(" · 忽略 %d 个空条目", ignored)
					}
					ui.Text(c, meta).FontSize(a.metrics.fontCaption).TextColor(p.text3)
				})
			})
			a.divider(c, false, p.line)
			ui.Column(c).Grow(0.88).Basis(0).FillHeight().Background(p.surface2).Children(func() {
				ui.Row(c).Height(a.metrics.paneHeaderH).FillWidth().Padding(0, space12).BorderWidth(0, 0, 1, 0).BorderColor(p.lineSoft).
					AlignItems(ui.Center).Gap(space6).Children(func() {
					ui.Text(c, "解析结果").FontSize(a.metrics.fontSmall).FontWeight(weightSemibold).TextColor(p.text2)
					a.chip(c, fmt.Sprint(len(entries)), p.text3, p.surface3)
					ui.Spacer(c).Grow(1)
					ui.Textf(c, "%s · %d → %d 项", b.target, len(info.original), len(entries)).FontSize(a.metrics.fontCaption).TextColor(p.text3).SingleLine()
				})
				ui.Scroll(c).Grow(1).Padding(space6, space4, space12, space6).Children(func() {
					a.batchStats(c, info)
					if len(entries) == 0 && len(info.original) > 0 {
						a.batchCallout(c, iconAlert, p.dangerSoft, p.danger, "应用后此变量的值为空。")
					} else if len(entries) == 0 {
						a.batchCallout(c, iconInfo, p.surface2, p.text3, "粘贴内容后，这里会显示解析出的条目。")
					}
					if info.reordered {
						a.batchCallout(c, iconInfo, p.warnSoft, p.warn, "条目顺序已调整，PATH 会按新顺序查找程序。")
					}
					if len(info.duplicates) > 0 {
						uniq := uniqueStrings(info.duplicates)
						a.batchCallout(c, iconAlert, p.warnSoft, p.warn,
							fmt.Sprintf("%d 个重复条目 · %s。重复项通常不会再次生效。", len(uniq), strings.Join(firstN(uniq, 2), "、")))
					}
					if len(entries) > 0 {
						ui.Textf(c, "新值条目  %d", len(entries)).FontSize(a.metrics.fontCaption).FontWeight(weightSemibold).
							LetterSpacing(0.5).TextColor(p.text3).Padding(space2, space4)
						for i, e := range entries {
							isNew := !containsFold(info.original, e)
							a.batchEntry(c, i+1, max(len(entries), len(info.removed)), e, isNew, false)
						}
					}
					if len(info.removed) > 0 {
						ui.Textf(c, "将移除  %d", len(info.removed)).FontSize(a.metrics.fontCaption).FontWeight(weightSemibold).
							LetterSpacing(0.5).TextColor(p.text3).Padding(space8, space4, space2, space4)
						for i, e := range info.removed {
							a.batchEntry(c, i+1, max(len(entries), len(info.removed)), e, false, true)
						}
					}
				})
			})
		})
	})
	if a.modal != nil {
		return
	}
	c.OnShortcut(0, ui.KeyEscape, func() { a.closeBatch() })
	c.OnShortcut(ui.Cmd, ui.KeyEnter, func() { a.applyBatch(c) })
	c.OnShortcut(ui.Cmd, ui.KeyS, func() { a.applyBatch(c) })
}

func (a *App) batchStats(c *ui.Context, info batchInfo) {
	p := a.pal
	ui.Row(c).FillWidth().Gap(space4).Padding(space2, space4, space6, space4).Children(func() {
		if len(info.added) > 0 {
			a.statChip(c, fmt.Sprintf("+%d 新增", len(info.added)), iconPlus, p.success, p.successSoft)
		}
		if len(info.removed) > 0 {
			a.statChip(c, fmt.Sprintf("−%d 移除", len(info.removed)), iconMinus, p.danger, p.dangerSoft)
		}
		if info.reordered {
			a.statChip(c, "顺序调整", iconUnfold, p.warn, p.warnSoft)
		}
		if !info.hasChanges {
			a.statChip(c, "与当前值一致", iconCheck, p.success, p.successSoft)
		}
	})
}

func (a *App) statChip(c *ui.Context, label string, ic *ui.SVG, fg, bg ui.Color) {
	ui.Row(c).Padding(space2, space4).Radius(radiusControl).Background(bg).Gap(space2).AlignItems(ui.Center).Children(func() {
		ui.Icon(c, ic).Size(9, 9).TextColor(fg)
		ui.Text(c, label).FontSize(a.metrics.fontCaption).FontWeight(weightSemibold).TextColor(fg)
	})
}

func (a *App) batchCallout(c *ui.Context, ic *ui.SVG, bg, fg ui.Color, text string) {
	ui.Row(c).FillWidth().Margin(0, space4, space6, space4).Padding(space6, space8).Radius(radiusControl).Background(bg).Gap(space6).AlignItems(ui.Start).Children(func() {
		ui.Icon(c, ic).Size(10, 10).TextColor(fg)
		ui.Text(c, text).FontSize(a.metrics.fontCaption).TextColor(fg).MaxLines(4)
	})
}

func (a *App) batchEntry(c *ui.Context, index, count int, text string, added, removed bool) {
	p := a.pal
	bg := p.surface
	border := p.line
	fg := p.text2
	mark := ""
	if added {
		bg = p.successSoft
		border = p.successSoft
		fg = p.success
		mark = "+"
	}
	if removed {
		bg = p.dangerSoft
		border = p.dangerSoft
		fg = p.danger
		mark = "−"
	}
	ui.Row(c).FillWidth().Margin(0, 0, space2, 0).Padding(space4, space8).Radius(radiusControl).Background(bg).Border(1, border).
		AlignItems(ui.Start).Gap(space8).Children(func() {
		a.listNumber(c, index, count)
		// Reserve the status column even for unchanged entries.
		ui.Text(c, mark).Font(monoFont).FontSize(a.metrics.fontSmall).TextColor(fg).Width(a.metrics.fontSmall).TextAlign(ui.Center)
		ui.Text(c, text).Font(monoFont).FontSize(a.metrics.fontSmall).TextColor(fg).Grow(1).MinWidth(0)
	})
}

// batchEditorTextWidth returns how wide the batch editor's text pane is: as
// wide as its widest line, but filling the viewport when that fits.
func (a *App) batchEditorTextWidth(c *ui.Context, text string) float32 {
	viewport, _ := a.batchEditorViewport(c, text)
	widest, widestRunes := "", 0
	for line := range strings.SplitSeq(text, "\n") {
		if n := utf8.RuneCountInString(line); n > widestRunes {
			widest, widestRunes = line, n
		}
	}
	if widest == "" {
		return viewport
	}
	w, _ := c.MeasureText(0, ui.Span{Text: widest, Font: monoFont, Size: a.metrics.fontBody})
	if w += 12; w < viewport {
		return viewport
	}
	return w
}

// batchEditorViewport estimates the space the batch editor's text pane gets,
// from the window's size and the measurements of the pane around it, so that
// the pane fills it: the text pane scrolls sideways only as long as a line
// is wider than this, and is as tall as this at least.
func (a *App) batchEditorViewport(c *ui.Context, text string) (width, height float32) {
	w, h := c.Size()
	titleBar := c.TitleBar().Height
	if titleBar < a.metrics.titleBarH {
		titleBar = a.metrics.titleBarH
	}
	left := (w - batchDividerWidth) * batchLeftFlex / (batchLeftFlex + batchRightFlex)
	width = left - 2*batchBodyPadH - 2*batchBoxBorder - 2*batchBoxPadH - numberColumnWidth(c, strings.Count(text, "\n")+1, a.metrics.fontBody) - batchGutterGap
	height = h - titleBar - a.metrics.pageHeaderH - a.metrics.paneHeaderH - a.metrics.barH -
		2*batchBodyPadV - 2*batchBoxBorder - 2*batchBoxPadV
	if width < 120 {
		width = 120
	}
	if floor := 3 * a.metrics.lineBody; height < floor {
		height = floor
	}
	return width, height
}

func (a *App) applyBatch(c *ui.Context) {
	b := a.batch
	if b == nil {
		return
	}
	entries, _ := batchParse(b.text)
	info := analyzeBatch(b.originalValue, b.text)
	if !info.hasChanges {
		return
	}
	if len(entries) == 0 && len(info.original) > 0 {
		a.modal = &modalState{
			icon:    iconAlert,
			title:   "清空变量值？",
			message: fmt.Sprintf("将把 “%s” 的值清空。应用前会自动备份，可在 5 秒内撤销。", b.target),
			confirm: "清空并应用",
			cancel:  "取消",
			danger:  true,
			onConfirm: func() {
				a.applyBatchValue(b.scope, b.target, "")
			},
		}
		return
	}
	value := strings.Join(entries, ";")
	a.applyBatchValue(b.scope, b.target, value)
}
