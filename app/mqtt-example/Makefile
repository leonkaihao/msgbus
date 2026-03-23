.PHONY: all

build:
	CGO_ENABLED=0 go build -o mqtt-example *.go
start_server:
	docker run -it -p 1883:1883 -v "$(PWD)/mosquitto:/mosquitto" amd64/eclipse-mosquitto