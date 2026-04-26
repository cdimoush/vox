//go:build ui

// Package level opens the default capture device via malgo and streams
// RMS audio levels to a channel at ~30 FPS.
//
// The mic is opened independently of SoX — modern audio stacks
// (PipeWire, PulseAudio, CoreAudio) allow concurrent capture. If the
// two conflict, the level monitor gracefully degrades (logs and stops).
package level

import (
	"context"
	"encoding/binary"
	"log"
	"math"
	"time"

	"github.com/gen2brain/malgo"
)

const (
	sampleRate = 44100
	// Target ~30 updates/sec. Each callback covers periodSize frames.
	periodMS = 33
)

// Monitor opens the default mic and sends RMS values (0.0–1.0) to the
// returned channel until ctx is cancelled. The channel is closed on
// exit. Errors are logged, not returned — the overlay works without
// levels.
func Monitor(ctx context.Context) <-chan float64 {
	ch := make(chan float64, 4)

	go func() {
		defer close(ch)
		if err := run(ctx, ch); err != nil {
			log.Printf("level monitor: %v", err)
		}
	}()

	return ch
}

func run(ctx context.Context, out chan<- float64) error {
	mctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}
	defer mctx.Free()

	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = 1
	cfg.SampleRate = sampleRate
	cfg.PeriodSizeInMilliseconds = periodMS

	// Throttle: send at most one value per ~30ms.
	var lastSend time.Time

	callbacks := malgo.DeviceCallbacks{
		Data: func(_, pInput []byte, frameCount uint32) {
			if len(pInput) == 0 {
				return
			}
			rms := rmsS16(pInput, frameCount)

			now := time.Now()
			if now.Sub(lastSend) < 30*time.Millisecond {
				return
			}
			lastSend = now

			select {
			case out <- rms:
			default:
				// drop if overlay is slow
			}
		},
	}

	dev, err := malgo.InitDevice(mctx.Context, cfg, callbacks)
	if err != nil {
		return err
	}
	defer dev.Uninit()

	if err := dev.Start(); err != nil {
		return err
	}
	log.Println("level monitor: capturing mic levels")

	<-ctx.Done()

	_ = dev.Stop()
	return nil
}

// rmsS16 computes the root-mean-square of signed 16-bit PCM samples,
// normalised to 0.0–1.0.
func rmsS16(buf []byte, frames uint32) float64 {
	n := int(frames)
	if n == 0 {
		return 0
	}
	// Each frame is 2 bytes (mono S16).
	var sumSq float64
	for i := 0; i < n && i*2+1 < len(buf); i++ {
		sample := int16(binary.LittleEndian.Uint16(buf[i*2:]))
		f := float64(sample) / 32768.0
		sumSq += f * f
	}
	rms := math.Sqrt(sumSq / float64(n))
	// Clamp to 1.0 (shouldn't exceed, but float rounding).
	if rms > 1 {
		rms = 1
	}
	return rms
}
