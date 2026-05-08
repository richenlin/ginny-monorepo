package graceful

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	closers             []func(ctx context.Context) error
	mainGoroutinePrefix = []byte("goroutine 1 ")
	signals             = make(chan os.Signal, 10)
)

// Fn fn is a function with error
type Fn func() error

// AddCloser add closer.
func AddCloser(closer func(ctx context.Context) error) {
	closers = append(closers, closer)
}

// Close close the app gracefully.
func Close() {
	signals <- nil
}

// Start start the app, if fn is not empty, it will start fn async
func Start(fn ...Fn) {
	for _, v := range fn {
		go func(f Fn) {
			err := f()
			if err != nil {
				warnJSONLog(err.Error())
				signals <- nil
			}
		}(v)
	}
	// check if in main thread
	var buf [16]byte
	n := runtime.Stack(buf[:], false)
	mainGoroutine := bytes.HasPrefix(buf[:n], mainGoroutinePrefix)

	if mainGoroutine {
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		if sig := <-signals; sig != nil {
			warnJSONLog(fmt.Sprintf("closed by %s", sig.String()))
		}
		runClosers()
	}
}

// runClosers close all registered closers.
func runClosers() {
	// Close all closers in the order of first-in-last-out.
	l := len(closers)
	for i := l - 1; i > -1; i-- {
		ctx, cc := context.WithTimeout(context.Background(), 6*time.Minute)
		err := closers[i](ctx)
		cc()
		if err != nil {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				if strings.HasPrefix(pathErr.Path, "/dev/std") {
					continue
				}
			}
			warnJSONLog(err.Error())
		}
	}
}

func warnJSONLog(msg string) {
	_, _ = fmt.Fprintf(os.Stdout, `{"level":"warn","time":"%s","msg":"%s"}`+"\n",
		time.Now().UTC().Format(time.RFC3339), msg)
}
