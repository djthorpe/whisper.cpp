package main

import (
	"context"
	"fmt"

	// Packages
	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	sdl "github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SDL struct {
	id       sdl.AudioDeviceID
	rate     int32
	channels uint8
	samples  uint16
	data     []byte
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSDL(samples uint16) (*SDL, error) {
	this := new(SDL)

	// Init audio
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		return nil, err
	} else {
		this.samples = samples
	}

	// Set priority
	sdl.LogSetPriority(sdl.LOG_CATEGORY_APPLICATION, sdl.LOG_PRIORITY_INFO)
	sdl.SetHintWithPriority(sdl.HINT_AUDIO_RESAMPLING_MODE, "medium", sdl.HINT_OVERRIDE)

	// Return success
	return this, nil
}

func (dev *SDL) Close() error {
	if dev.id != 0 {
		sdl.CloseAudioDevice(dev.id)
	}
	sdl.Quit()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (dev *SDL) String() string {
	str := "<sdl"
	if dev.rate != 0 {
		str += fmt.Sprintf(" rate=%v", dev.rate)
	}
	if dev.channels != 0 {
		str += fmt.Sprintf(" channels=%v", dev.channels)
	}
	if dev.samples != 0 {
		str += fmt.Sprintf(" samples=%v", dev.samples)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (dev *SDL) NumDevices() int {
	return sdl.GetNumAudioDevices(true)
}

func (dev *SDL) DeviceName(i int) string {
	return sdl.GetAudioDeviceName(i, true)
}

func (dev *SDL) Open(num int) error {
	var requested, obtained sdl.AudioSpec
	requested.Freq = whisper.SampleRate
	requested.Format = sdl.AUDIO_F32
	requested.Channels = 1
	requested.Samples = dev.samples
	if id, err := sdl.OpenAudioDevice(sdl.GetAudioDeviceName(num, true), true, &requested, &obtained, 0); err != nil {
		return err
	} else {
		dev.id = id
	}

	// Set parameters
	dev.rate = obtained.Freq
	dev.channels = obtained.Channels
	dev.samples = obtained.Samples
	dev.data = make([]byte, dev.samples*(whisper.SampleBits>>3))

	// Return success
	return nil
}

func (dev *SDL) Rate() int {
	return int(dev.rate)
}

func (dev *SDL) BitDepth() int {
	return int(whisper.SampleBits)
}

func (dev *SDL) Channels() int {
	return int(dev.channels)
}

func (dev *SDL) Capture(ctx context.Context, fn func()) error {
	if dev.id == 0 {
		return fmt.Errorf("no device open")
	}

	// Capture on/off
	sdl.PauseAudioDevice(dev.id, false)
	defer sdl.PauseAudioDevice(dev.id, true)

	// Capture loop until cancelled
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := sdl.DequeueAudio(dev.id, dev.data); err != nil {
				return err
			} else if fn != nil {
				fn()
			}
			queued := sdl.GetQueuedAudioSize(dev.id)
			if queued > 0 {
				fmt.Println("Still queued=", queued)
			}
		}
	}
}

func (dev *SDL) Bytes() []byte {
	return dev.data
}
