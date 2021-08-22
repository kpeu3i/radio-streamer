package streaming

import (
	"fmt"
	"strconv"
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
	Volume() float64
	SetVolume(v float64)
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

func (s *Service) PlayRadio() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.radioPlayer.IsPlaying() {
		return nil
	}

	config, err := s.configStorage.Load()
	if err != nil {
		return err
	}

	volume, err := strconv.ParseFloat(config.CurrentVolume, 64)
	if err != nil {
		return err
	}

	s.radioPlayer.SetVolume(volume)
	s.radioPlayer.Play(config.CurrentStream)

	return nil
}

func (s *Service) StopRadio() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.radioPlayer.IsPlaying() {
		return
	}

	s.radioPlayer.Stop()
}

func (s *Service) IsRadioPlaying() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.radioPlayer.IsPlaying()
}

func (s *Service) PrevRadioStream() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	num := s.radioPlayer.Prev()

	config, err := s.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentStream = num

	err = s.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) NextRadioStream() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	num := s.radioPlayer.Next()

	config, err := s.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentStream = num

	err = s.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpVolume(step float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	volume := s.radioPlayer.Volume()

	volume += step
	if volume > 1 {
		volume = 1
	}

	s.radioPlayer.SetVolume(volume)

	config, err := s.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentVolume = fmt.Sprintf("%.2f", volume)

	err = s.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DownVolume(step float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	volume := s.radioPlayer.Volume()

	volume -= step
	if volume < 0 {
		volume = 0
	}

	s.radioPlayer.SetVolume(volume)

	config, err := s.configStorage.Load()
	if err != nil {
		return err
	}

	config.CurrentVolume = fmt.Sprintf("%.2f", volume)

	err = s.configStorage.Store(config)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.radioPlayer.IsPlaying() {
		return nil
	}

	return s.radioPlayer.Close()
}
