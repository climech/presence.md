package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"presence/app"
	"presence/config"
	"presence/server"
	"syscall"
)

const APPNAME = "presence"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		dief("couldn't load home: %v", err)
	}

	dir := filepath.Join(home, ".config", APPNAME)
	conf, err := config.LoadConfig(dir)
	if err != nil {
		dief("couldn't load config: %v", err)
	}

	a, err := app.New(conf)
	if err != nil {
		die(err)
	}
	defer a.Close()

	s, err := server.New(a)
	if err != nil {
		die(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		fmt.Println("\nshutting down...")
		s.Close()
	}()

	if err := s.Run(); err != nil {
		die(err)
	}
}
