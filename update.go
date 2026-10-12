package main

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
)

// Updates is the update section of the settings page.
type Updates struct {
	Enabled            bool      `json:"enabled"` // false in development builds
	AutomaticChecks    bool      `json:"automaticChecks"`
	AutomaticDownloads bool      `json:"automaticDownloads"`
	LastCheck          time.Time `json:"lastCheck"`
}

// Updates returns the update preferences.
func (l *Library) Updates() Updates {
	return Updates{
		Enabled:            mygo.Updater.Enabled(),
		AutomaticChecks:    updater.AutomaticChecks(),
		AutomaticDownloads: updater.AutomaticDownloads(),
		LastCheck:          updater.LastCheck(),
	}
}

// SetUpdates changes the update preferences.
func (l *Library) SetUpdates(checks, downloads bool) Updates {
	updater.SetAutomaticChecks(checks)
	updater.SetAutomaticDownloads(downloads)
	return l.Updates()
}

// CheckForUpdates shows the update window, which says when the app is up to
// date.
func (l *Library) CheckForUpdates() {
	updater.CheckForUpdates()
}
