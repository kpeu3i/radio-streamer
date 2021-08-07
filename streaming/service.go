package streaming

import (
	"sync"

	"github.com/kpeu3i/radio-streamer/radio"
)

type ConfigStorage interface {
	Load() (Config, error)
	Store(config Config) error
}

type RadioPlayer interface {
	Play(streamNum int)
	Prev() int
	Next() int
	IsPlaying() bool
	OnError(handler radio.ErrorHandler)
	Volume() (int, error)
	SetVolume(v int) error
	Stop()
	Close() error
}

type Service struct {
	configStorage ConfigStorage
	radioPlayer   RadioPlayer
	mu            sync.Mutex
}

func NewService(configStorage ConfigStorage, radioPlayer RadioPlayer) *Service {
	return &Service{configStorage: configStorage, radioPlayer: radioPlayer}
}

func (a *Service) PlayRadio() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.radioPlayer.IsPlaying() {
		return nil
	}

	config, err := a.configStorage.Load()
	if err != nil {
		return err
	}

	err = a.radioPlayer.SetVolume(config.CurrentVolume)
	if err != nil {
		return err
	}

	a.radioPlayer.Play(config.CurrentStream)

	return nil
}

func (a *Service) StopRadio() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.radioPlayer.IsPlaying() {
		return
	}

	a.radioPlayer.Stop()
}

func (a *Service) IsRadioPlaying() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.radioPlayer.IsPlaying()
}

func (a *Service) PrevRadioStream() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	num := a.radioPlayer.Prev()

	config, err := a.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentStream = num

	err = a.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (a *Service) NextRadioStream() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	num := a.radioPlayer.Next()

	config, err := a.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentStream = num

	err = a.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (a *Service) UpVolume(step int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	volume, err := a.radioPlayer.Volume()
	if err != nil {
		return err
	}

	volume += step
	if volume > 100 {
		volume = 100
	}

	err = a.radioPlayer.SetVolume(volume)
	if err != nil {
		return err
	}

	config, err := a.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentVolume = volume

	err = a.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (a *Service) DownVolume(step int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	volume, err := a.radioPlayer.Volume()
	if err != nil {
		return err
	}

	volume -= step
	if volume < 0 {
		volume = 0
	}

	err = a.radioPlayer.SetVolume(volume)
	if err != nil {
		return err
	}

	config, err := a.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentVolume = volume

	err = a.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (a *Service) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.radioPlayer.IsPlaying() {
		return nil
	}

	return a.radioPlayer.Close()
}
