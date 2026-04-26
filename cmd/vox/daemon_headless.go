//go:build !ui

package main

import (
	"github.com/cdimoush/vox/daemon"
	"github.com/cdimoush/vox/daemon/config"
)

// startUISubsystems is a no-op when built without the ui tag.
// The daemon still works — hotkey triggering is via `vox ui toggle` (IPC).
func startUISubsystems(_ *daemon.Daemon, _ config.Config, _ chan struct{}) {}
