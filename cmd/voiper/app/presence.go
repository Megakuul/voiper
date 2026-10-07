package app

func (a *App) UnpublishStatus(account string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.phone.UnpublishStatus(account)
}

func (a *App) UnwatchPresence(account, target string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.phone.Unwatch(account, target)
}
