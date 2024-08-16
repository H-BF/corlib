package observer

type (
	//EventType тип сообщения
	EventType interface {
		isObserverEventType()
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
