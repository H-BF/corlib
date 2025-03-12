package observer

type (
	//EventType тип сообщения
	EventType interface {
		isObserverEventType()
	}

	// EventTypeBaseImpl impl EventType iface
	EventTypeBaseImpl struct{}

	// EventTypeT тип сообщения + свойство
	EventTypeT[T any] struct {
		EventTypeBaseImpl
		Property T
	}

	//EventReceiver получалель сообщений
	EventReceiver func(event EventType)

	//Observer тот кто получит сообщения
	Observer interface {
		Close() error
		SubscribeEvents(...EventType)
		UnsubscribeEvents(...EventType)
		UnsubscribeAllEvents()
		Observe(...EventType)
	}

	//Subject источник сообщений
	Subject interface {
		ObserversAttach(...Observer)
		ObserversDetach(...Observer)
		DetachAllObservers()
		Notify(...EventType)
	}
)

func (EventTypeBaseImpl) isObserverEventType() {} // impl EventType

// EventTypeOf -
func EventTypeOf[T any](arg T) EventTypeT[T] {
	return EventTypeT[T]{
		Property: arg,
	}
}
