package radio

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hajimehoshi/oto/v2"
	"github.com/tosone/minimp3"
)

const (
	contextSampleRate  = 44100
	contextNumChannels = 2
)

type ErrorHandler func(err error)

type Player struct {
	streams      []string
	stream       io.ReadCloser
	decoder      *minimp3.Decoder
	context      *oto.Context
	player       oto.Player
	volume       float64
	errorHandler ErrorHandler
	index        int
	play         chan string
	stop         chan struct{}
}

func NewPlayer(streams ...string) *Player {
	return &Player{
		streams: streams,
		volume:  1,
		play:    make(chan string),
		stop:    make(chan struct{}),
		errorHandler: func(err error) {
			log.Printf("An error occured while playing/stopping: %s\n", err)
		},
	}
}

func (p *Player) Play(streamNum int) {
	if len(p.streams) == 0 {
		return
	}

	if p.IsPlaying() {
		return
	}

	index := streamNum - 1
	if index < 0 || index >= len(p.streams) {
		index = 0
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if v, ok := r.(error); ok {
					p.errorHandler(v)
				} else {
					p.errorHandler(fmt.Errorf("%v", v))
				}
			}
		}()

		err := p.run(index)
		if err != nil {
			p.errorHandler(err)
		}
	}()

	return
}

func (p *Player) Prev() int {
	if !p.IsPlaying() {
		return p.index + 1
	}

	p.play <- p.prevStream()

	return p.index + 1
}

func (p *Player) Next() int {
	if !p.IsPlaying() {
		return p.index + 1
	}

	p.play <- p.nextStream()

	return p.index + 1
}

func (p *Player) IsPlaying() bool {
	return p.player != nil && p.player.IsPlaying()
}

func (p *Player) OnError(handler ErrorHandler) {
	p.errorHandler = handler
}

func (p *Player) Volume() float64 {
	if p.IsPlaying() {
		return p.player.Volume()
	}

	return p.volume
}

func (p *Player) SetVolume(v float64) {
	if p.IsPlaying() {
		p.player.SetVolume(v)
	}

	p.volume = v
}

func (p *Player) Stop() {
	if !p.IsPlaying() {
		return
	}

	p.stop <- struct{}{}

	err := p.Close()
	if err != nil {
		p.errorHandler(err)
	}
}

func (p *Player) Close() error {
	return p.free()
}

func (p *Player) doPlay(stream string) error {
	response, err := http.Get(stream)
	if err != nil {
		return err
	}

	decoder, err := minimp3.NewDecoder(response.Body)
	if err != nil {
		return err
	}

	if isStarted := <-decoder.Started(); !isStarted {
		return errors.New("cannot start decoding")
	}

	// TODO https://github.com/hajimehoshi/oto/issues/149
	if p.context == nil {
		context, ready, err := oto.NewContext(contextSampleRate, contextNumChannels, 2)
		if err != nil {
			return err
		}

		<-ready

		p.context = context
	}

	p.stream = response.Body
	p.decoder = decoder
	p.player = p.context.NewPlayer(decoder)

	log.Printf(
		"Audio stream initialized (url: %s, bitrate: %d, samplerate: %d, channels: %d)\n",
		stream,
		p.decoder.Kbps,
		p.decoder.SampleRate,
		p.decoder.Channels,
	)

	p.player.SetVolume(p.volume)
	p.player.Play()

	return nil
}

func (p *Player) run(index int) error {
	err := p.doPlay(p.exactStream(index))
	if err != nil {
		return err
	}

	for {
		select {
		case url := <-p.play:
			err := p.free()
			if err != nil {
				return err
			}

			err = p.doPlay(url)
			if err != nil {
				return err
			}

		case <-p.stop:
			return nil
		}
	}
}

func (p *Player) free() error {
	err := p.player.Close()
	if err != nil {
		return err
	}

	p.decoder.Close()

	err = p.stream.Close()
	if err != nil {
		return err
	}

	p.stream = nil
	p.decoder = nil
	// TODO https://github.com/hajimehoshi/oto/issues/149
	// p.context = nil
	p.player = nil

	return nil
}

func (p *Player) currentStream() string {
	return p.streams[p.index]
}

func (p *Player) exactStream(index int) string {
	p.index = index

	return p.streams[p.index]
}

func (p *Player) prevStream() string {
	p.index--
	if p.index < 0 {
		p.index = len(p.streams) - 1
	}

	return p.streams[p.index]
}

func (p *Player) nextStream() string {
	p.index++
	if p.index > len(p.streams)-1 {
		p.index = 0
	}

	return p.streams[p.index]
}
