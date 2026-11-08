package app

import "testing"

func TestSheetPreservesValueAcrossModes(t *testing.T) {
	for _, value := range []string{"a;;b;", " a ; b ", "", ";;", "a\r\nb;c"} {
		t.Run(value, func(t *testing.T) {
			a, _ := newTestApp(t)
			a.openEdit(Variable{Name: "CUSTOM", Value: value}, true)
			if a.sheetDirty() || a.sheet.currentValue() != value {
				t.Fatalf("opening changed %q to %q", value, a.sheet.currentValue())
			}
			for _, mode := range []int{0, 1, 0, 1} {
				a.sheet.valueMode = mode
				a.sheet.syncValueMode()
				if a.sheetDirty() || a.sheet.currentValue() != value {
					t.Fatalf("mode %d changed value to %q", mode, a.sheet.currentValue())
				}
			}
			a.closeSheet()
			if a.sheet != nil || a.modal != nil {
				t.Fatal("clean editor requested discard confirmation")
			}
		})
	}
}

func TestSheetListSavePreservesUntouchedItems(t *testing.T) {
	a, tt := newTestApp(t)
	backend := memoryBackend(t, a)
	backend.Replace(ScopeUser, []Variable{{Name: "CUSTOM", Value: " a ;;b;", Type: "REG_SZ"}})
	a.reload()
	v, _ := a.findVariable(ScopeUser, "CUSTOM")
	a.openEdit(v, true)
	a.saveSheet(nil)
	if len(a.snapshots) != 0 {
		t.Fatal("unchanged save created backup")
	}
	a.sheet.entries[2] = "changed"
	tt.Frame()
	if err := tt.Click("保存更改"); err != nil {
		t.Fatal(err)
	}
	got, _ := a.findVariable(ScopeUser, "CUSTOM")
	if got.Value != " a ;;changed;" {
		t.Fatalf("untouched items changed: %q", got.Value)
	}
}

func TestCloseGuardsUnsavedEditors(t *testing.T) {
	for _, mode := range []string{"sheet", "batch"} {
		t.Run(mode, func(t *testing.T) {
			a, tt := newTestApp(t)
			if mode == "sheet" {
				a.openCreate("UNSAVED")
				a.sheet.value = "draft"
			} else {
				a.openBatch("Path")
				a.batch.text = "UNSAVED"
			}
			closed := false
			closeWindow := func() { closed = true }
			if a.allowClose(closeWindow) {
				t.Fatal("dirty editor allowed close")
			}
			tt.Frame()
			if err := tt.Click("继续编辑"); err != nil {
				t.Fatal(err)
			}
			if closed || (!a.sheetDirty() && !a.batchDirty()) {
				t.Fatal("cancel lost edits")
			}
			if a.allowClose(closeWindow) {
				t.Fatal("second close skipped confirmation")
			}
			tt.Frame()
			if err := tt.Click("放弃并退出"); err != nil {
				t.Fatal(err)
			}
			if !closed || a.sheet != nil || a.batch != nil {
				t.Fatal("discard did not close")
			}
			if !a.allowClose(closeWindow) {
				t.Fatal("clean close prompted again")
			}
		})
	}
}

func TestBatchTargetSwitchRetainsDraftUntilConfirmed(t *testing.T) {
	a, _ := newTestApp(t)
	a.openBatch("Path")
	a.batch.text = "draft"
	a.switchBatchTarget("PYTHONPATH")
	if a.batch.target != "Path" || a.batch.text != "draft" || a.modal == nil {
		t.Fatal("switch lost draft before confirmation")
	}
	a.dismissModal(false)
	if a.batch.target != "Path" || a.batch.text != "draft" {
		t.Fatal("cancel lost draft")
	}
	a.switchBatchTarget("PYTHONPATH")
	a.dismissModal(true)
	if a.batch.target != "PYTHONPATH" || a.batchDirty() {
		t.Fatal("confirmed switch did not load target")
	}
}
