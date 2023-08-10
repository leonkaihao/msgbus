package model

type Service interface {
	// AddConsumers create and manage routine for each consumer
	AddConsumers(cms []Consumer)
	// Serve is blocked util Close() is invoked.
	Serve() error
	// Close release the resource it used
	Close() error
}

type ServiceCallback interface {
	OnReceive(csmr Consumer, msg Messager)
	OnDestroy(inst interface{})
	OnError(inst interface{}, err error)
}
