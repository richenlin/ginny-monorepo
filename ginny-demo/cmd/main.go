package main

import (
	"context"
	"time"

	"github.com/goriller/ginny/logger"
	_ "go.uber.org/automaxprocs/maxprocs"
)

func main() {
	ctx, cc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cc()
	app, err := NewApp(ctx)
	if err != nil {
		logger.Action("NewApp").Fatal(err.Error())
	}
	if err := app.Start(ctx); err != nil {
		logger.Action("Start").Fatal(err.Error())
	}
}
