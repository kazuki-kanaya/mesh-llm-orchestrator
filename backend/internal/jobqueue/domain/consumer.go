package domain

import "errors"

type MessageID string
type ConsumerName string

var (
	ErrInvalidMessageID    = errors.New("invalid message id")
	ErrInvalidConsumerName = errors.New("invalid consumer name")
)

func (id MessageID) String() string {
	return string(id)
}

func (id MessageID) Validate() error {
	if id == "" {
		return ErrInvalidMessageID
	}
	return nil
}

func (name ConsumerName) String() string {
	return string(name)
}

func (name ConsumerName) Validate() error {
	if name == "" {
		return ErrInvalidConsumerName
	}
	return nil
}
