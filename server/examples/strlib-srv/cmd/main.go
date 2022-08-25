package main

import (
	"context"
	"time"

	"github.com/H-BF/corlib/logger"
	pkgNet "github.com/H-BF/corlib/pkg/net"
	"github.com/H-BF/corlib/pkg/signals"
	"github.com/H-BF/corlib/server"
	strlibSrv "github.com/H-BF/corlib/server/examples/strlib-srv"
	"github.com/H-BF/corlib/server/interceptors"
	serverMetrics "github.com/H-BF/corlib/server/metrics/prometheus"
	"go.uber.org/zap"
)

func main() {
	logger.SetLevel(zap.InfoLevel)

	//1 - создаем контекст приложения
	ctx, cancel := context.WithCancel(context.Background())
	signals.WhenSignalExit(func() error { //2 - канцеляем контекст если получили сигнал "навыход"
		cancel()
		return nil
	})

	logger.Info(ctx, "--== hello ==--")

	strlibService := strlibSrv.NewStrLibService(ctx) //3 - создаем сервис с API
	docs, err := strlibService.GetDocs()             //4 - если есть свагер - доки = достаем их
	if err != nil {
		logger.Fatal(ctx, err)
	}

	var endpoint *pkgNet.Endpoint
	if endpoint, err = pkgNet.ParseEndpoint("tcp://127.0.0.1:8002"); err != nil { //5 - будем чалить сервер в этом адресе
		logger.Fatal(ctx, err)
	}

	var strlibServer *server.APIServer
	//6 - создаем GRPC / GW сервер со свагером + API
	metrics := serverMetrics.NewMetrics(serverMetrics.WithNamespace("sample"))
	serverRecovery := interceptors.NewRecovery(interceptors.RecoveryWithObservers(metrics.PanicsObserver()))
	strlibServer, err = server.NewAPIServer(server.WithServices(strlibService),
		server.WithDocs(docs, ""),
		server.WithStatsHandlers(metrics.StatHandlers()...), //подключаем  метрики
		server.WithRecovery(serverRecovery))                 //подключаем RECOVERY+метрики
	if err != nil {
		logger.Fatal(ctx, err)
	}

	//7 - запускаем сервер | в браузере http://127.0.0.1:8002/docs - покажет SwaggerUI
	err = strlibServer.Run(ctx, endpoint, server.RunWithGracefulStop(30*time.Second))
	if err != nil {
		logger.Fatal(ctx, err)
	}
	logger.Info(ctx, "--== bye ==--")

}
