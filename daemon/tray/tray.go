//go:build ui

// Package tray provides a system-tray icon that reflects the daemon's
// state and offers a menu for basic control (toggle, settings, quit).
//
// On Linux, this requires AppIndicator support. GNOME users need the
// AppIndicator extension. If the tray fails to start, the daemon still
// works — the tray is best-effort.
package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/systray"

	"github.com/cdimoush/vox/daemon"
	"github.com/cdimoush/vox/daemon/config"
)

// icon colors per state.
var (
	colorIdle       = color.NRGBA{R: 140, G: 140, B: 140, A: 255}
	colorRecording  = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
	colorTranscribe = color.NRGBA{R: 50, G: 120, B: 220, A: 255}
	colorDone       = color.NRGBA{R: 40, G: 160, B: 80, A: 255}
)

// Run starts the system tray and blocks until the daemon shuts down.
// Call from a goroutine. quit is closed to signal daemon shutdown
// (same channel used by IPC).
func Run(d *daemon.Daemon, quit chan struct{}) {
	events := d.Subscribe()

	// Pre-render icons.
	icons := map[daemon.State][]byte{
		daemon.StateIdle:         circleIcon(colorIdle, 22),
		daemon.StateRecording:    circleIcon(colorRecording, 22),
		daemon.StateTranscribing: circleIcon(colorTranscribe, 22),
		daemon.StateDone:         circleIcon(colorDone, 22),
	}

	var mToggle, mSettings, mQuit *systray.MenuItem

	onReady := func() {
		systray.SetIcon(icons[daemon.StateIdle])
		systray.SetTooltip("vox — idle")

		mToggle = systray.AddMenuItem("Toggle Recording", "Start or stop recording")
		systray.AddSeparator()
		mSettings = systray.AddMenuItem("Settings", "Open ~/.vox/config.toml")
		systray.AddSeparator()
		mQuit = systray.AddMenuItem("Quit", "Stop vox daemon")

		go handleEvents(d, events, icons, mToggle, mSettings, mQuit, quit)
	}

	onExit := func() {
		log.Println("tray: exited")
	}

	// RunWithExternalLoop lets the tray coexist with Gio's event loop.
	start, end := systray.RunWithExternalLoop(onReady, onExit)
	start()
	log.Println("tray: started")

	// Block until quit is signalled.
	<-quit
	systray.Quit()
	end()
}

func handleEvents(
	d *daemon.Daemon,
	events <-chan daemon.Event,
	icons map[daemon.State][]byte,
	mToggle, mSettings, mQuit *systray.MenuItem,
	quit chan struct{},
) {
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return
			}
			if icon, exists := icons[ev.To]; exists {
				systray.SetIcon(icon)
			}
			systray.SetTooltip("vox — " + ev.To.String())

		case <-mToggle.ClickedCh:
			d.Toggle()

		case <-mSettings.ClickedCh:
			openConfig()

		case <-mQuit.ClickedCh:
			select {
			case <-quit:
				// already closed
			default:
				close(quit)
			}
			return

		case <-quit:
			return
		}
	}
}

// openConfig opens ~/.vox/config.toml in the user's preferred editor.
func openConfig() {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		log.Printf("tray: config path: %v", err)
		return
	}
	// Ensure the file exists so the editor has something to open.
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(cfgPath), 0o755)
		_ = os.WriteFile(cfgPath, []byte("# vox daemon config — see docs/design/vox-ui/system-design.md §8\n"), 0o600)
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "xdg-open"
	}
	cmd := exec.Command(editor, cfgPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("tray: open config: %v", err)
	}
}

// circleIcon renders a filled circle as a PNG byte slice. Simple but
// functional — replace with real assets once the UI is validated.
func circleIcon(c color.NRGBA, size int) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	center := float64(size) / 2
	rSq := center * center
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center + 0.5
			dy := float64(y) - center + 0.5
			if dx*dx+dy*dy <= rSq {
				img.SetNRGBA(x, y, c)
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
