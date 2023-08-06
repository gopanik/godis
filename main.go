package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gopanik/godis/internal"
	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	addr := ":6379"
	s := internal.NewServer(addr, logger)
	err := s.ListenAndServe()
	if err != nil {
		logger.Error("Error listening to tcp", slog.String("addr", addr), slog.String("err", err.Error()))
		os.Exit(1)
	}

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGINT, syscall.SIGTERM)

	<-cancelChan
	s.Stop()
}
