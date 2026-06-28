package events

import "github.com/shagston/routerpilot/sdk/types"

type Publisher interface {
	Publish(types.Event) error
}

type Subscriber interface {
	Handle(types.Event) error
}

type Bus interface {
	Publisher
	Subscribe(Subscriber) error
}
