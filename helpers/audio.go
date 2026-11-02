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

const maxSoundPlayers = 4

type soundPool struct {
	pcm     []byte
	players []*audio.Player
	next    int
}

type SoundManager struct {
	context *audio.Context
	sounds  map[string][]byte
	pools   map[string]*soundPool
	muted   bool
}

func NewSoundManager(assets embed.FS, soundFiles map[string]string) *SoundManager {
	manager := &SoundManager{
		sounds: loadSounds(assets, soundFiles),
		pools:  make(map[string]*soundPool, len(soundFiles)),
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

	pool := s.pools[name]
	if pool == nil {
		pool = &soundPool{pcm: pcm}
		s.pools[name] = pool
	}

	player := pool.idlePlayer(s.context)
	player.SetVolume(volume)
	_ = player.Rewind()
	player.Play()
}

func (p *soundPool) idlePlayer(context *audio.Context) *audio.Player {
	for i := range p.players {
		index := (p.next + i) % len(p.players)
		if !p.players[index].IsPlaying() {
			p.next = (index + 1) % len(p.players)
			return p.players[index]
		}
	}

	if len(p.players) < maxSoundPlayers {
		player := context.NewPlayerFromBytes(p.pcm)
		p.players = append(p.players, player)
		p.next = len(p.players) % maxSoundPlayers
		return player
	}

	player := p.players[p.next]
	p.next = (p.next + 1) % len(p.players)
	return player
}
