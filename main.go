package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kpeu3i/radio-streamer/httpapi"
	"github.com/kpeu3i/radio-streamer/mqttapi"
	"github.com/kpeu3i/radio-streamer/radio"
	"github.com/kpeu3i/radio-streamer/streaming"
)

const (
	configFilepath = "config.yaml"

	appRestartIntervalHours = 3
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	appConfig, err := newConfig()
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	configStorage := streaming.NewConfigStorage(configFilePath())
	streamingServiceConfig, err := configStorage.Load()
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	errs := make(chan error)
	wasPlaying := false

	for {
		radioPlayer := radio.NewPlayer(streamingServiceConfig.Streams...)
		radioPlayer.OnError(func(err error) { errs <- err })
		httpServer := httpapi.NewServer(appConfig.HTTPServer.Address)
		mqttListener := mqttapi.NewListener(
			appConfig.MQTTServer.Address,
			appConfig.MQTTServer.User,
			appConfig.MQTTServer.Password,
			appConfig.MQTTServer.Topic,
		)
		service := streaming.NewService(configStorage, radioPlayer)
		panicHandler := func(v interface{}) { errs <- fmt.Errorf("%v", v) }

		if wasPlaying {
			err = service.PlayRadio()
			if err != nil {
				log.Fatalf("[ERROR] %v", err)
			}
		}

		now := time.Now()
		nextRestartDuration := nextTickDuration(now.Hour()+appRestartIntervalHours, now.Minute(), now.Second())
		restartTimer := time.NewTimer(nextRestartDuration)

		go func() {
			log.Println("Starting application...")
			log.Printf("Scheduled application restart time: %s", now.Add(nextRestartDuration).Format(time.RFC3339))

			err := runApp(httpServer, mqttListener, service, panicHandler)
			if err != nil {
				errs <- err
			}
		}()

		select {
		case <-signals:
			log.Println("Got termination signal")
			log.Println("Stopping application...")

			_ = stopApp(httpServer, mqttListener, service)

			return
		case <-restartTimer.C:
			log.Println("Got restart signal")
			log.Println("Stopping application...")

			wasPlaying = service.IsRadioPlaying()
			restartTimer.Stop()
			_ = stopApp(httpServer, mqttListener, service)
			time.Sleep(appConfig.ErrorHandling.RecoveryDelay)
		case err = <-errs:
			if err == nil {
				return
			}

			log.Printf("Got runtime error: %v", err)
			log.Fatalf("[ERROR] %v", err)
		}
	}
}

func runApp(
	httpServer *httpapi.Server,
	mqttListener *mqttapi.Listener,
	service *streaming.Service,
	panicHandler func(v interface{}),
) error {
	errs := make(chan error)

	go func() {
		errs <- runHTTPServer(httpServer, service, panicHandler)
	}()

	go func() {
		errs <- runMQTTServer(mqttListener, service, panicHandler)
	}()

	return <-errs
}

func stopApp(httpServer *httpapi.Server, mqttListener *mqttapi.Listener, service *streaming.Service) []error {
	var errs []error

	err := httpServer.Close()
	if err != nil {
		errs = append(errs, err)
	}

	err = mqttListener.Close()
	if err != nil {
		errs = append(errs, err)
	}

	err = service.Close()
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func runHTTPServer(
	httpServer *httpapi.Server,
	service *streaming.Service,
	panicHandler func(v interface{}),
) error {
	httpServer.
		Register("/radio/power", httpapi.WrapHandler(
			httpapi.RadioPowerHandler(service),
			httpapi.RecoverMiddleware(panicHandler),
		)).
		Register("/radio/stream/prev", httpapi.WrapHandler(
			httpapi.RadioStreamPrevHandler(service),
			httpapi.RecoverMiddleware(panicHandler),
		)).
		Register("/radio/stream/next", httpapi.WrapHandler(
			httpapi.RadioStreamNextHandler(service),
			httpapi.RecoverMiddleware(panicHandler),
		)).
		Register("/radio/volume/up", httpapi.WrapHandler(
			httpapi.VolumeUpHandler(service),
			httpapi.RecoverMiddleware(panicHandler),
		)).
		Register("/radio/volume/down", httpapi.WrapHandler(
			httpapi.VolumeDownHandler(service),
			httpapi.RecoverMiddleware(panicHandler),
		))

	return httpServer.Listen()
}

func runMQTTServer(
	mqttListener *mqttapi.Listener,
	service *streaming.Service,
	panicHandler func(v interface{}),
) error {
	mqttListener.
		Register("button_1_click", mqttapi.WrapHandler(
			mqttapi.RadioPowerHandler(service),
			mqttapi.RecoverMiddleware(panicHandler),
		)).
		Register("button_2_click", mqttapi.WrapHandler(
			mqttapi.RadioStreamNextHandler(service),
			mqttapi.RecoverMiddleware(panicHandler),
		)).
		Register("button_2_hold", mqttapi.WrapHandler(
			mqttapi.RadioStreamPrevHandler(service),
			mqttapi.RecoverMiddleware(panicHandler),
		)).
		Register("button_3_click", mqttapi.WrapHandler(
			mqttapi.VolumeDownHandler(service),
			mqttapi.RecoverMiddleware(panicHandler),
		)).
		Register("button_4_click", mqttapi.WrapHandler(
			mqttapi.VolumeUpHandler(service),
			mqttapi.RecoverMiddleware(panicHandler),
		))

	return mqttListener.Listen()
}

func configFilePath() string {
	ex, _ := os.Executable()

	return path.Join(filepath.Dir(ex), configFilepath)
}

func nextTickDuration(hour, min, sec int) time.Duration {
	now := time.Now()

	nextTick := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		hour,
		min,
		sec,
		0,
		time.Local,
	)

	if nextTick.Before(now) {
		nextTick = nextTick.Add(24 * time.Hour)
	}

	return nextTick.Sub(now)
}
