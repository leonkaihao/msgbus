
$(eval RELEASE_TAG := $(shell jq -r '.release_tag' component.json))
$(eval VERSION := $(shell jq -r '.version' component.json))
$(eval REGISTRY := $(shell jq -r '.registry' component.json))

.PHONY: all build

build:
	go mod tidy && go build ./pkg/client/nats
test:
	go mod tidy && go test -v -cover ./pkg/model/...
test/integration:
	go mod tidy && go test -v --tags=integration ./pkg/model/...
test/bench:
	go mod tidy && go test -bench=. -benchmem ./pkg/model/...
test/pack:
	go mod tidy && go test -c -o=bin/platform.msgbus.test ./pkg/model/...
	go test -c -o=bin/platform.msgbus.integrationtest --tags=integration ./pkg/model/...
clean:
	rm -rf ./bin
	go clean -modcache


build-consumer:
	docker build -t sim/nats/consumer:$(RELEASE_TAG) -f ./build/docker/Dockerfile_consumer .
	docker tag sim/nats/consumer:$(RELEASE_TAG) $(REGISTRY)/nats-consumer:$(RELEASE_TAG)
build-producer:
	docker build -t sim/nats/producer:$(RELEASE_TAG) -f ./build/docker/Dockerfile_producer .
	docker tag sim/nats/producer:$(RELEASE_TAG) $(REGISTRY)/nats-producer:$(RELEASE_TAG)
push-consumer:
	docker push $(REGISTRY)/nats-consumer:$(RELEASE_TAG)
	docker tag $(REGISTRY)/nats-consumer:$(RELEASE_TAG) $(REGISTRY)/nats-consumer:$(VERSION)-${BUILD_NUMBER}
	docker push $(REGISTRY)/nats-consumer:$(VERSION)-${BUILD_NUMBER}
push-producer:
	docker push $(REGISTRY)/nats-producer:$(RELEASE_TAG)
	docker tag $(REGISTRY)/nats-producer:$(RELEASE_TAG) $(REGISTRY)/nats-producer:$(VERSION)-${BUILD_NUMBER}
	docker push $(REGISTRY)/nats-producer:$(VERSION)-${BUILD_NUMBER}

