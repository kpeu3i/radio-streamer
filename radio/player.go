package radio

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hajimehoshi/oto"
	"github.com/itchyny/volume-go"
	"github.com/tosone/minimp3"
)

type ErrorHandler func(err error)

type Player struct {
	streams      []string
	stream       io.ReadCloser
	decoder      *minimp3.Decoder
	context      *oto.Context
	player       *oto.Player
	isPlaying    bool
	errorHandler ErrorHandler
	index        int
	play         chan string
	stop         chan struct{}
}

func NewPlayer(streams ...string) *Player {
	return &Player{
		streams: streams,
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

	if p.isPlaying {
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
	if !p.isPlaying {
		return p.index + 1
	}

	p.play <- p.prevStream()

	return p.index + 1
}

func (p *Player) Next() int {
	if !p.isPlaying {
		return p.index + 1
	}

	p.play <- p.nextStream()

	return p.index + 1
}

func (p *Player) IsPlaying() bool {
	return p.isPlaying
}

func (p *Player) OnError(handler ErrorHandler) {
	p.errorHandler = handler
}

func (p *Player) Volume() (int, error) {
	return volume.GetVolume()
}

func (p *Player) SetVolume(v int) error {
	return volume.SetVolume(v)
}

func (p *Player) Stop() {
	if !p.isPlaying {
		return
	}

	p.stop <- struct{}{}

	err := p.Close()
	if err != nil {
		p.errorHandler(err)
	}
}

func (p *Player) Close() error {
	err := p.player.Close()
	if err != nil {
		return err
	}

	err = p.context.Close()
	if err != nil {
		log.Printf("Failed to close context: %s\n", err)
	}

	p.decoder.Close()

	err = p.stream.Close()
	if err != nil {
		return err
	}

	p.stream = nil
	p.decoder = nil
	p.context = nil
	p.player = nil

	return nil
}

func (p *Player) init(stream string) error {
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

	context, err := oto.NewContext(decoder.SampleRate, decoder.Channels, 2, 4096)
	if err != nil {
		return err
	}

	p.stream = response.Body
	p.decoder = decoder
	p.context = context
	p.player = context.NewPlayer()

	log.Printf(
		"Audio stream initialized (url: %s, bitrate: %d, samplerate: %d, channels: %d)\n",
		stream,
		p.decoder.Kbps,
		p.decoder.SampleRate,
		p.decoder.Channels,
	)

	return nil
}

func (p *Player) run(index int) error {
	err := p.init(p.exactStream(index))
	if err != nil {
		return err
	}

	p.isPlaying = true
	defer func() {
		p.isPlaying = false
	}()

	for {
		select {
		case url := <-p.play:
			err := p.Close()
			if err != nil {
				return err
			}

			err = p.init(url)
			if err != nil {
				return err
			}
		case <-p.stop:
			return nil
		default:
			data := make([]byte, 512)
			_, err := p.decoder.Read(data)
			if err != nil {
				if err == io.EOF {
					break
				}

				return err
			}

			_, err = p.player.Write(data)
			if err != nil {
				return err
			}
		}
	}
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
