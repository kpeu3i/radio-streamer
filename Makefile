include .env

SHELL := /bin/sh
.DEFAULT_GOAL := default
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CURRENT_DIR := $(patsubst %/,%,$(dir $(MAKEFILE_PATH)))

DOCKER_MOUNT_DIR := /radio-streamer
DOCKER_IMAGE := radio-streamer
DOCKER_RUN := docker run -it --rm --entrypoint="" \
	-v ${CURRENT_DIR}:${DOCKER_MOUNT_DIR} \
	-v ~/go/pkg/mod:/go/pkg/mod \
	-w ${DOCKER_MOUNT_DIR} \
	${DOCKER_IMAGE}

.PHONY: docker-build
docker-build:
	@docker build -f docker/Dockerfile -t ${DOCKER_IMAGE} ./docker

.PHONY: compile
compile:
	@echo "Compiling..."
	@${DOCKER_RUN} bash -c ' \
		CC=arm-linux-gnueabihf-gcc-5 \
		GOOS=linux \
		GOARCH=arm \
		GOARM=7 \
		CGO_ENABLED=1 \
		go build \
	'

.PHONY: upload
upload:
	@ssh ${REMOTE_HOST} mkdir -p /home/pi/radio-streamer
	@ssh ${REMOTE_HOST} rm -f /home/pi/radio-streamer/radio-streamer
	@scp ${CURRENT_DIR}/radio-streamer ${REMOTE_HOST}:/home/pi/radio-streamer/radio-streamer
	@scp ${CURRENT_DIR}/config.yaml ${REMOTE_HOST}:/home/pi/radio-streamer/config.yaml

.PHONY: deploy
deploy: compile service-stop upload service-start

.PHONY: service-init
service-init:
	@scp ${CURRENT_DIR}/resources/radio-streamer.service ${REMOTE_HOST}:/home/pi/radio-streamer/radio-streamer.service
	@ssh ${REMOTE_HOST} sudo cp /home/pi/radio-streamer/radio-streamer.service /etc/systemd/system/radio-streamer.service
	@ssh ${REMOTE_HOST} rm -f /home/pi/radio-streamer/radio-streamer.service
	@ssh ${REMOTE_HOST} sudo systemctl daemon-reload
	@ssh ${REMOTE_HOST} systemctl enable radio-streamer
	@ssh ${REMOTE_HOST} sudo systemctl start radio-streamer

.PHONY: service-start
service-start:
	@ssh ${REMOTE_HOST} "sudo service radio-streamer start"

.PHONY: service-stop
service-stop:
	@ssh ${REMOTE_HOST} "sudo service radio-streamer stop"

.PHONY: service-restart
service-restart:
	@ssh ${REMOTE_HOST} "sudo service radio-streamer restart"

.PHONY: service-status
service-status:
	@ssh ${REMOTE_HOST} "sudo service radio-streamer status"

.PHONY: service-logs
service-logs:
	@ssh ${REMOTE_HOST} "journalctl -u radio-streamer"

.PHONY: ssh
ssh:
	@ssh ${REMOTE_HOST}
