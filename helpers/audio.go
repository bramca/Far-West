package helpers

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	audioSampleRate = 44100
	soundDirectory  = "assets/sfx/"
)

var (
	audioContextOnce sync.Once
	audioContext     *audio.Context
)

type SoundManager struct {
	context *audio.Context
	sounds  map[string][]byte
	muted   bool
}

func NewSoundManager(assets embed.FS, soundFiles map[string]string) *SoundManager {
	manager := &SoundManager{
		sounds: loadSounds(assets, soundFiles),
		muted:  testing.Testing(),
	}
	if !manager.muted {
		manager.context = getAudioContext()
	}

	return manager
}

func getAudioContext() *audio.Context {
	audioContextOnce.Do(func() {
		audioContext = audio.NewContext(audioSampleRate)
	})

	return audioContext
}

func loadSounds(assets embed.FS, soundFiles map[string]string) map[string][]byte {
	sounds := make(map[string][]byte, len(soundFiles))
	for name, fileName := range soundFiles {
		raw, err := assets.ReadFile(soundDirectory + fileName)
		if err != nil {
			panic(fmt.Sprintf("could not read sound %q: %v", name, err))
		}

		stream, err := wav.DecodeWithSampleRate(audioSampleRate, bytes.NewReader(raw))
		if err != nil {
			panic(fmt.Sprintf("could not decode sound %q: %v", name, err))
		}

		pcm, err := io.ReadAll(stream)
		if err != nil {
			panic(fmt.Sprintf("could not load sound %q: %v", name, err))
		}

		sounds[name] = pcm
	}

	return sounds
}

func (s *SoundManager) Play(name string, volume float64) {
	if s == nil || s.muted {
		return
	}

	pcm, ok := s.sounds[name]
	if !ok {
		return
	}

	player := s.context.NewPlayerFromBytes(pcm)
	player.SetVolume(volume)
	player.Play()
}
