package inproc

import "errors"

var (
	ErrEmptySubConsumer = errors.New("number of topic consumers = 0, the instance is not avaiable")
)
