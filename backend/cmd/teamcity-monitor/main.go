package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nightnoryu/go-kita/env"
	"github.com/nightnoryu/go-kita/jsonlog"
	"github.com/nightnoryu/go-kita/log"
)

const appID = "teamcity_monitor"

func main() {
	ctx := context.Background()
	logger := initLogger()
	defer func() { _ = logger.Sync() }()

	cnf, err := env.ParseEnv[config](appID)
	if err != nil {
		logger.FatalError(err)
	}

	err = runApp(ctx, cnf, logger)
	if errors.Is(err, errServiceStopped) {
		logger.Info(err.Error())
	} else {
		logger.FatalError(err)
	}
}

func runApp(ctx context.Context, config *config, logger log.Logger) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	ctx = listenOSKillSignals(ctx)

	if len(os.Args) != 2 {
		return errors.New("mode argument not provided")
	}

	mode := os.Args[1]
	if mode == "service" {
		return service(ctx, config, logger)
	}
	return fmt.Errorf("unknown mode: %s", mode)
}

func initLogger() log.MainLogger {
	return jsonlog.NewLogger(&jsonlog.Config{
		Level:   jsonlog.InfoLevel,
		AppName: appID,
	})
}

func listenOSKillSignals(ctx context.Context) context.Context {
	var cancelFunc context.CancelFunc
	ctx, cancelFunc = context.WithCancel(ctx)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-ch:
			cancelFunc()
		case <-ctx.Done():
			signal.Reset()
			return
		}
	}()
	return ctx
}
