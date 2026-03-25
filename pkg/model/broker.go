package model

type Broker interface {
	// URL endpoint to message broker
	URL() string
	// Consumer generate a consumer from this broker
	Consumer(name string, topic string, group string) Consumer
	// Producer generate a consumer from this broker
	Producer(name string, topic string) Producer
	// release the resource of its own
	Close() error
	// RemoveConsumer delete consumer from management
	RemoveConsumer(csmr Consumer) error
	// RemoveProducer delete producer from management
	RemoveProducer(prd Producer) error
}
