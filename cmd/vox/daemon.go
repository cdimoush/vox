package main

import "errors"

// cmdDaemon is the hidden `vox __daemon` entrypoint. The real loop lives
// in daemon_main.go (landed in task .6). This placeholder keeps the build
// green while task .5 (vox ui CLI) lands ahead of it.
func cmdDaemon() error {
	return errors.New("vox __daemon: not yet implemented (use vox ui start after task .6 lands)")
}
