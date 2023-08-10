package model

type Client interface {
	// Broker generates a Broker by an url endpoint
	Broker(url string) Broker
	// Remove a Broker from its management
	Remove(brk Broker) error
	// close all the brokers
	Close() error
}
