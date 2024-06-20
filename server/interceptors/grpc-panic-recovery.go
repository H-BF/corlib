package interceptors

import (
	"context"
	"reflect"
	"runtime"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/pkg/patterns/observer"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	//RecoveryOption ...
	RecoveryOption func(*recoveryOptions)

	//OnPanicEvent такая штука должнп рассылаться всем кто подписался на нее
	OnPanicEvent struct {
		observer.EventType
		Info       conventions.GrpcMethodInfo
		PanicCause interface{}
		Ctx        context.Context
	}

	//OnPanicEventObserver ...
	OnPanicEventObserver func(OnPanicEvent)

	//CustomRecoveryHandler custom recovery from panic error handler
	CustomRecoveryHandler func(ctx context.Context, info conventions.GrpcMethodInfo, v interface{}) error
)

// RecoveryWithNoLog не логируем панику
func RecoveryWithNoLog() RecoveryOption {
	return func(options *recoveryOptions) {
		options.noLog = true
	}
}

// RecoveryWithNoCallStack не добавляем стек вызова при логоровании
func RecoveryWithNoCallStack() RecoveryOption {
	return func(options *recoveryOptions) {
		options.noCallStack = true
	}
}

// RecoveryWithObservers добавим OnPanicEvent обозревателей
func RecoveryWithObservers(obs ...OnPanicEventObserver) RecoveryOption {
	return func(options *recoveryOptions) {
		options.observers = append(options.observers, obs...)
	}
}

// RecoveryWithHandler добавим кастомный обработчик паники
func RecoveryWithHandler(h CustomRecoveryHandler) RecoveryOption {
	return func(options *recoveryOptions) {
		options.customHandler = h
	}
}

// Recovery ...
type Recovery struct {
	noRecovery bool
	opts       recoveryOptions
	subject    observer.Subject
}

// NewRecovery делаем обработчик паники
func NewRecovery(opts ...RecoveryOption) *Recovery {
	ret := new(Recovery)
	for _, o := range opts {
		o(&ret.opts)
	}
	if len(ret.opts.observers) > 0 {
		ret.subject = observer.NewSubject()
		var evt OnPanicEvent
		seen := make(map[reflect.Value]bool)
		for _, obs := range ret.opts.observers {
			if v := reflect.ValueOf(obs); !seen[v] {
				seen[v] = true
			} else {
				continue
			}
			o := observer.NewObserver(func(event observer.EventType) {
				if ev, ok := event.(OnPanicEvent); ok {
					obs(ev)
				}
			}, false, evt)
			ret.subject.ObserversAttach(o)
		}
		ret.opts.observers = nil
		runtime.SetFinalizer(ret, func(o *Recovery) {
			o.subject.DetachAllObservers()
		})
	}
	return ret
}

// NoRecovery никиаких обработчикао паники - паника - роняем приложение
func NoRecovery() *Recovery {
	return &Recovery{
		noRecovery: true,
	}
}

type recoveryOptions struct {
	customHandler CustomRecoveryHandler
	observers     []OnPanicEventObserver
	noLog         bool
	noCallStack   bool
}

// Unary ...
func (impl *Recovery) Unary(ctx context.Context, req interface{}, i *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if !impl.noRecovery {
		defer func() {
			if p := recover(); p != nil {
				err = impl.handleAfterRecovery(ctx, i.FullMethod, p)
			}
		}()
	}
	resp, err = handler(ctx, req)
	return
}

// Stream stream server interceptor
func (impl *Recovery) Stream(srv interface{}, ss grpc.ServerStream, i *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	if !impl.noRecovery {
		defer func(ctx context.Context) {
			if p := recover(); p != nil {
				err = impl.handleAfterRecovery(ctx, i.FullMethod, p)
			}
		}(ss.Context())
	}
	err = handler(srv, ss)
	return
}

func (impl *Recovery) handleAfterRecovery(ctx context.Context, method string, p interface{}) error {
	if p == nil {
		return nil
	}
	var info conventions.GrpcMethodInfo
	if !info.FromContext(ctx) {
		e := info.Init(method)
		if e != nil {
			panic(e)
		}
	}
	var err error
	if f := impl.opts.customHandler; f != nil {
		err = f(ctx, info, p)
	}
	if err == nil {
		err = status.Errorf(codes.Internal, "PANIC(%v)", p)
	} else {
		err = status.Errorf(codes.Internal, "PANIC(%v)", err)
	}
	err1 := err
	if !impl.opts.noCallStack {
		err1 = errors.WithStack(err)
	}
	if !impl.opts.noLog && err1 != nil {
		st, _ := err1.(interface {
			StackTrace() errors.StackTrace
		})
		kv := []interface{}{
			"cause", err1.Error(),
			"service", info.ServiceFQN,
			"method", info.Method,
		}
		if st != nil {
			kv = append(kv, "stack_trace", st.StackTrace())
		}
		logger.ErrorKV(ctx, "PANIC-RECOVERY", kv...)
	}
	impl.notifyAboutRecovery(ctx, info, p)
	return err
}

func (impl *Recovery) notifyAboutRecovery(ctx context.Context, inf conventions.GrpcMethodInfo, p interface{}) {
	if subj := impl.subject; subj != nil {
		subj.Notify(OnPanicEvent{
			Info:       inf,
			PanicCause: p,
			Ctx:        ctx,
		})
	}
}
