package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nightnoryu/go-kita/env"
	"github.com/nightnoryu/go-kita/jsonlog"
	"github.com/nightnoryu/go-kita/log"
)

const appID = "teamcity_monitor"

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	logger, err := initLogger()
	if err != nil {
		stdlog.Fatal(err)
	}
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
	if len(os.Args) != 2 {
		return errors.New("mode argument not provided")
	}

	mode := os.Args[1]
	if mode == "service" {
		return service(ctx, config, logger)
	}
	return fmt.Errorf("unknown mode: %s", mode)
}

func initLogger() (log.MainLogger, error) {
	return jsonlog.NewLogger(&jsonlog.Config{
		Level:   jsonlog.InfoLevel,
		AppName: appID,
	})
}
