package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"

	"github.com/archproj/slackoverflow/config"
	"github.com/archproj/slackoverflow/database"
	"github.com/archproj/slackoverflow/listen"
	m "github.com/archproj/slackoverflow/middlewares"
	"github.com/archproj/slackoverflow/slack"
)

const (
	VERSION = "0.2.0"
)

func main() {
	cfg, err := config.Load() // from environment
	if err != nil {
		log.Panic(err)
	}

	e := echo.New()

	db, err := database.Init(cfg)
	if err != nil {
		log.Panic(err)
	}

	sc, err := slack.Init(cfg)
	if err != nil {
		log.Panic(err)
	}

	e.Use(m.EmbedInContext(cfg, db, sc))

        e.POST("/listen/command", listen.CommandHandler)

	go func() {
		err := e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
		if err != nil {
			log.Warning("Shutting down server...")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
