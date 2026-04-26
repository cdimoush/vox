package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

// cmdUIInstall writes a platform-appropriate autostart unit so the vox
// daemon launches on login. It does NOT start the daemon — run
// `vox ui start` or re-login after install.
func cmdUIInstall() error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate self: %w", err)
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	switch runtime.GOOS {
	case "linux":
		return installSystemd(self)
	case "darwin":
		return installLaunchd(self)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// cmdUIUninstall removes the autostart unit.
func cmdUIUninstall() error {
	switch runtime.GOOS {
	case "linux":
		return uninstallSystemd()
	case "darwin":
		return uninstallLaunchd()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// --- systemd (Linux) -----------------------------------------------------

const systemdUnit = `[Unit]
Description=Vox dictation daemon
Documentation=https://github.com/cdimoush/vox
After=graphical-session.target

[Service]
Type=simple
ExecStart={{.Binary}} __daemon
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

func systemdUnitPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "systemd", "user", "vox.service"), nil
}

func installSystemd(binary string) error {
	path, err := systemdUnitPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("unit").Parse(systemdUnit))
	if err := tmpl.Execute(f, struct{ Binary string }{binary}); err != nil {
		return err
	}

	// Reload so systemd picks up the new unit.
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: systemctl daemon-reload failed: %v\n", err)
	}
	if err := exec.Command("systemctl", "--user", "enable", "vox.service").Run(); err != nil {
		return fmt.Errorf("enable unit: %w", err)
	}

	fmt.Fprintf(os.Stderr, "installed: %s\n", path)
	fmt.Fprintln(os.Stderr, "daemon will start on next login, or run: systemctl --user start vox")
	return nil
}

func uninstallSystemd() error {
	path, err := systemdUnitPath()
	if err != nil {
		return err
	}
	_ = exec.Command("systemctl", "--user", "stop", "vox.service").Run()
	_ = exec.Command("systemctl", "--user", "disable", "vox.service").Run()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	fmt.Fprintln(os.Stderr, "vox autostart removed")
	return nil
}

// --- launchd (macOS) -----------------------------------------------------

const launchdPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.vox.daemon</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.Binary}}</string>
    <string>__daemon</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key>
    <false/>
  </dict>
  <key>StandardOutPath</key>
  <string>{{.LogDir}}/daemon.log</string>
  <key>StandardErrorPath</key>
  <string>{{.LogDir}}/daemon.log</string>
</dict>
</plist>
`

func launchdPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", "com.vox.daemon.plist"), nil
}

func installLaunchd(binary string) error {
	path, err := launchdPlistPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".vox")

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("plist").Parse(launchdPlist))
	if err := tmpl.Execute(f, struct{ Binary, LogDir string }{binary, logDir}); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "installed: %s\n", path)
	fmt.Fprintln(os.Stderr, "daemon will start on next login, or run: launchctl load "+path)
	return nil
}

func uninstallLaunchd() error {
	path, err := launchdPlistPath()
	if err != nil {
		return err
	}
	_ = exec.Command("launchctl", "unload", path).Run()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	fmt.Fprintln(os.Stderr, "vox autostart removed")
	return nil
}
