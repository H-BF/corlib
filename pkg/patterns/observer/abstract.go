package observer

type (
	//EventType тип сообщения
	EventType interface {
		isObserverEventType()
	}
	// EventTypeT тип сообщения + свойство
	EventTypeT[T any] struct {
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

func (EventTypeT[T]) isObserverEventType() {} // impl EventType

// EventTypeOf -
func EventTypeOf[T any](arg T) (ret EventTypeT[T]) {
	ret.Property = arg
	return ret
}
