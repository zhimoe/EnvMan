package app

func (a *App) takeSnapshot(scope Scope, note string) (Snapshot, error) {
	return a.service.Snapshot(scope, note)
}
