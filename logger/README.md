Logger
------

This package provides wrapper for [Zap logger](https://github.com/uber-go/zap/) compatible with our [logger convention](https://confluence.ozon.ru/pages/viewpage.action?pageId=85830279).  
The most important feature is opentracing metadata injection to log records. If passed `context` contains span then all records will have `trace_id` & `span_id` fields.

## How to use?

There is default global logger that writes all logs (configured with DEBUG level).

For each level there are three functions that allows to format log record in various ways:

1. Simple logging

```go
logger.Error(ctx, "hello", "world")
// output: 
// {"level":"error","ts":"2018-12-21T13:14:05.747+0300","message":"helloworld","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
```

2. Formatted logging
```go
logger.Errorf(ctx, "hello: %v", "world")
// output:
// {"level":"error","ts":"2018-12-21T13:16:52.271+0300","message":"hello: world","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
```

3. Key-value logging
```go
logger.ErrorKV(ctx, "hello world", "foo", "bar", "x", "y")
// output:
// {"level":"error","ts":"2018-12-21T13:18:02.536+0300","message":"hello world","foo":"bar","x":"y","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
```

### Set global logger

Default log level is ERROR, if you want you can use another default log level:
```go
l := logger.New(zapcore.DebugLevel)
logger.SetLogger(l)

// ...

logger.Debug(ctx, "debug message")
logger.Warn(ctx, "warn message")

// output:
// {"level":"debug","ts":"2018-12-21T13:27:04.028+0300","message":"debug message","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
// {"level":"warn","ts":"2018-12-21T13:27:04.028+0300","message":"warn message","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
```

Both DEBUG & WARN messages will be written.

### context.Context integration

You can inject another logger to `Context` and pass it to `logger.Debug` func and that logger will be used instead of default one.

```go
l := logger.New(zapcore.DebugLevel)
ctx := logger.ToContext(context.Background(), l)

logger.Debug(ctx, "debug message")
logger.Warn(ctx, "warn message")
// output:
// {"level":"debug","ts":"2018-12-21T13:28:49.621+0300","message":"debug message""trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
// {"level":"warn","ts":"2018-12-21T13:28:49.621+0300","message":"warn message""trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
```

### Add `caller`

If you want to add caller file & line to each log record you can add option to `New` func
```go
l := logger.New(
    zapcore.DebugLevel,
    zap.AddCaller(),
    zap.AddCallerSkip(1), // <-- required too
)

ctx := logger.ToContext(context.Background(), l)

logger.Debug(ctx, "debug message")
logger.Warn(ctx, "warn message")

// output:
// {"level":"debug","ts":"2018-12-21T13:31:15.272+0300","caller":"example/main.go:25","message":"debug message","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
// {"level":"warn","ts":"2018-12-21T13:31:15.272+0300","caller":"example/main.go:26","message":"warn message","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
```

### Add `stacktrace`

Stacktrace can be automatically added to records starting from some level.
```go
l := logger.New(
    zapcore.DebugLevel,
    zap.AddStacktrace(zapcore.ErrorLevel),
)

ctx := logger.ToContext(context.Background(), l)

logger.Debug(ctx, "debug message")
logger.Warn(ctx, "warn message")
logger.Error(ctx, "error message")

// output: 
// {"level":"debug","ts":"2018-12-21T13:37:12.511+0300","message":"debug message","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
// {"level":"warn","ts":"2018-12-21T13:37:12.511+0300","message":"warn message","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
// {"level":"error","ts":"2018-12-21T13:37:12.511+0300","message":"error message","stacktrace":"---stacktrace--here---","trace_id":"1a93a60e17efa0bd","span_id":"7cbc711c34a1f44d"}
```

So only records will level ERROR and lower will have stacktrace.
