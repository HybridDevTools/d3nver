package main

import (
	"context"
	"denver/cmd"
	"denver/cmd/root"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if cmd.WorkingDirectory == "" {
		cmd.WorkingDirectory = filepath.Dir(ex)
	}

	ctx, cancel := context.WithCancel(context.Background())

	rootApp := root.New(ctx, cmd.WorkingDirectory)

	go func() {
		<-sigs
		cancel()
		os.Exit(1)
	}()

	res := rootApp.Execute()
	cancel()
	os.Exit(res)
}
