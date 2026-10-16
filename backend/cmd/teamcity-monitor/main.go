package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nightnoryu/go-kita/jsonlog"
	"github.com/nightnoryu/go-kita/log"
)

const appID = "teamcity_monitor"

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFunc()

	cnf, err := loadConfig()
	if err != nil {
		stdlog.Fatal(err)
	}
	logger, err := initLogger(cnf.LogLevel)
	if err != nil {
		stdlog.Fatal(err)
	}
	defer func() { _ = logger.Sync() }()

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

func initLogger(level jsonlog.Level) (log.MainLogger, error) {
	return jsonlog.NewLogger(&jsonlog.Config{
		Level:   level,
		AppName: appID,
	})
}
