package interceptors

import (
	"context"
	"net/http"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/conventions"
	"github.com/H-BF/corlib/server/internal"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//LogLevelOverrider ...
var LogLevelOverrider logLvlOverride

type logLvlOverride struct{}

//HTTP is a http.Handler wrapper
func (logLvlOverride) HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var level zapcore.Level
		if lvl := r.Header.Get(conventions.LoggerLevelHeader); lvl != "" {
			currentLevel := logger.Level()
			if err := level.Set(lvl); err == nil && level != currentLevel {
				ctx := r.Context()
				oldLogger := logger.FromContext(ctx)
				newLogger := logger.TypeOfLogger{
					LevelEnabler: level,
					SugaredLogger: oldLogger.
						Desugar().
						WithOptions(logger.WithLevel(level)).
						Sugar(),
				}
				ctx = logger.ToContext(ctx, newLogger)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})

}

//Unary overrides unary interceptor log level
func (logLvlOverride) Unary(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		lvl := md.Get(conventions.LoggerLevelHeader)
		if len(lvl) > 0 {
			var level zapcore.Level
			currentLevel := logger.Level()
			if err := level.Set(lvl[0]); err == nil && level != currentLevel {
				oldLogger := logger.FromContext(ctx)
				newLogger := logger.TypeOfLogger{
					LevelEnabler: level,
					SugaredLogger: oldLogger.
						Desugar().
						WithOptions(logger.WithLevel(level)).
						Sugar(),
				}
				ctx = logger.ToContext(ctx, newLogger)
			}
		}
	}
	return handler(ctx, req)
}

//Stream overrides stream interceptor log level
func (logLvlOverride) Stream(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
		lvl := md.Get(conventions.LoggerLevelHeader)
		if len(lvl) > 0 {
			var level zapcore.Level
			currentLevel := logger.Level()
			if err := level.Set(lvl[0]); err == nil && level != currentLevel {
				ctx := ss.Context()
				oldLogger := logger.FromContext(ctx)
				newLogger := logger.TypeOfLogger{
					LevelEnabler: level,
					SugaredLogger: oldLogger.
						Desugar().
						WithOptions(logger.WithLevel(level)).
						Sugar(),
				}
				ctx = logger.ToContext(ctx, newLogger)
				ss = internal.ServerStreamWithContext(ctx, ss)
			}
		}
	}
	return handler(srv, ss)
}
