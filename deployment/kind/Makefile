.PHONY: all
all: | push-all deploy-all
push-all: push-consumer push-producer push-prometheus
deploy-all: | deploy-nats-stack deploy-nats-monitor

push-consumer:
	kind load docker-image sim/nats/consumer:1.0
push-producer:
	kind load docker-image sim/nats/producer:1.0
push-prometheus:
	kind load docker-image nats/prometheus:1.0

deploy-nats:
	kubectl apply -f ./nats.yaml
deploy-nats-svc:
	kubectl apply -f ./nats-svc.yaml
deploy-consumer:
	kubectl apply -f ./consumer.yaml
deploy-producer:
	kubectl apply -f ./producer.yaml
deploy-nats-reporter:
	kubectl apply -f ./nats-reporter.yaml
deploy-nats-reporter-svc:
	kubectl apply -f ./nats-reporter-svc.yaml
deploy-prometheus:
	kubectl apply -f ./prometheus.yaml
deploy-prometheus-svc:
	kubectl apply -f ./prometheus-svc.yaml
deploy-grafana:
	kubectl apply -f ./grafana.yaml

deploy-nats-stack: | deploy-nats deploy-nats-svc 
deploy-nats-monitor: | deploy-nats-reporter deploy-nats-reporter-svc deploy-prometheus deploy-prometheus-svc deploy-grafana
deploy-nats-simulator: deploy-consumer deploy-producer
