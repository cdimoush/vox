package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/cdimoush/vox/daemon/ipc"
)

// cmdUI dispatches `vox ui <subcmd>`.
func cmdUI() error {
	if len(os.Args) < 3 {
		return errors.New(uiUsage)
	}
	switch os.Args[2] {
	case "start":
		return cmdUIStart()
	case "stop":
		return cmdUIStop()
	case "toggle":
		return cmdUIToggle()
	case "status":
		return cmdUIStatus()
	case "install":
		return cmdUIInstall()
	case "uninstall":
		return cmdUIUninstall()
	default:
		return fmt.Errorf("unknown ui subcommand: %s\n\n%s", os.Args[2], uiUsage)
	}
}

const uiUsage = `Usage: vox ui {start|stop|toggle|status|install|uninstall}

  start     Launch the vox daemon in the background
  stop      Stop the running daemon
  toggle    Start or stop recording (for compositor hotkeys on Wayland)
  status    Show daemon state
  install   Set up daemon to start on login (systemd/launchd)
  uninstall Remove login autostart`

// pidPath returns ~/.vox/daemon.pid.
func pidPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".vox", "daemon.pid"), nil
}

// cmdUIStart forks the current binary as `vox __daemon`, detaches it,
// and records the PID. Returns once the daemon's socket is reachable
// (so follow-up `vox ui toggle` is race-free).
func cmdUIStart() error {
	pidFile, err := pidPath()
	if err != nil {
		return err
	}
	if pid, ok := readLivePID(pidFile); ok {
		return fmt.Errorf("vox daemon already running (pid %d)", pid)
	}
	if err := os.MkdirAll(filepath.Dir(pidFile), 0o755); err != nil {
		return err
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate self: %w", err)
	}

	logPath := filepath.Join(filepath.Dir(pidFile), "daemon.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open %s: %w", logPath, err)
	}
	defer logFile.Close()

	cmd := exec.Command(self, "__daemon")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("fork daemon: %w", err)
	}
	// Capture the PID before Release() — Release resets cmd.Process state.
	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return err
	}

	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0o600); err != nil {
		return fmt.Errorf("write pidfile: %w", err)
	}

	if err := waitForSocket(ipc.DefaultSocketPath(), 3*time.Second); err != nil {
		return fmt.Errorf("daemon did not come up: %w (see %s)", err, logPath)
	}

	fmt.Fprintf(os.Stderr, "vox daemon started (pid %d)\n", pid)
	return nil
}

// cmdUIStop sends shutdown and waits for the process to exit.
func cmdUIStop() error {
	pidFile, err := pidPath()
	if err != nil {
		return err
	}
	pid, running := readLivePID(pidFile)
	if !running {
		fmt.Fprintln(os.Stderr, "vox daemon is not running")
		_ = os.Remove(pidFile)
		return nil
	}

	if _, err := ipc.Dial(ipc.DefaultSocketPath(), ipc.Request{Cmd: ipc.CmdShutdown}); err != nil {
		// Socket may already be gone if the process is wedged — fall back
		// to SIGTERM so `vox ui stop` is reliable.
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}

	// Poll the PID for up to 3s.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err != nil {
			_ = os.Remove(pidFile)
			fmt.Fprintln(os.Stderr, "vox daemon stopped")
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("daemon (pid %d) did not exit in time", pid)
}

// cmdUIToggle sends a single toggle command.
func cmdUIToggle() error {
	resp, err := ipc.Dial(ipc.DefaultSocketPath(), ipc.Request{Cmd: ipc.CmdToggle})
	if err != nil {
		return fmt.Errorf("contact daemon: %w\n\nIs it running? Try: vox ui start", err)
	}
	if !resp.OK {
		return errors.New(resp.Err)
	}
	return nil
}

// cmdUIStatus reports daemon state, or "not running" if it's down.
func cmdUIStatus() error {
	pidFile, err := pidPath()
	if err != nil {
		return err
	}
	pid, running := readLivePID(pidFile)
	if !running {
		fmt.Println("not running")
		return nil
	}
	resp, err := ipc.Dial(ipc.DefaultSocketPath(), ipc.Request{Cmd: ipc.CmdStatus})
	if err != nil {
		fmt.Printf("pid %d (socket unreachable)\n", pid)
		return nil
	}
	fmt.Printf("pid %d, state=%s\n", pid, resp.State)
	return nil
}

// readLivePID returns (pid, true) if the pidfile exists and the process
// is alive, otherwise (0, false). A stale pidfile is not removed here —
// cmdUIStart and cmdUIStop handle cleanup.
func readLivePID(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil || pid <= 0 {
		return 0, false
	}
	// Signal 0 probes existence without actually signalling.
	if err := syscall.Kill(pid, 0); err != nil {
		return pid, false
	}
	return pid, true
}

// waitForSocket polls until the daemon's socket is accept-ready or the
// timeout elapses. Used by `vox ui start` to close the race between fork
// and the first `vox ui toggle`.
func waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := ipc.Dial(path, ipc.Request{Cmd: ipc.CmdStatus})
		if err == nil && resp.OK {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("timeout waiting for socket")
}

// marshalForTest exists so tests can round-trip request shapes without
// pulling in the whole ipc package. Not used at runtime.
var _ = json.Marshal
