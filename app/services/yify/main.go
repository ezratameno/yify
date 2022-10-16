package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"

	"github.com/ezratameno/yify/app/services/yify/handlers"
	"github.com/ezratameno/yify/business/sys/database"
	"github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "development"

func main() {
	log := logrus.New().WithField("service", "yify")
	if err := run(log); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(log *logrus.Entry) error {

	// ============================================================
	// GOMAXPROCS

	// Set the correct number of threads for the service
	// based on what is available either bt the machine or quotas.
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}

	// ============================================================
	// Configuration

	// default configuration

	cfg := struct {
		conf.Version
		Web struct {
			APIHost      string        `conf:"default:0.0.0.0:3000"`
			DebugHost    string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout  time.Duration `conf:"default:5s"`
			WriteTimeout time.Duration `conf:"default:10s"`
			IdleTimeout  time.Duration `conf:"default:120s"`
			ShutDownTime time.Duration `conf:"default:20s"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:localhost"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copyright information here",
		},
	}

	// prefix of the env vars to look for.
	const prefix = "YIFY"
	// ParseOSArgs - looks in the env, cli to override one or more of the flags
	help, err := conf.ParseOSArgs(prefix, &cfg)
	if err != nil {
		// if we run with the -h
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}
	// ============================================================
	// App Starting
	log.Infof("starting service, version: %s", cfg.Version)
	defer log.Info("shutdown complete")
	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info("startup", "config", out)
	log.Info("api host", cfg.Web.APIHost)

	// =========================================================================
	// Database Support

	// Create connectivity to the database.
	log.WithFields(logrus.Fields{
		"status": "initializing database support",
		"host":   cfg.DB.Host,
	}).Info("startup")

	db, err := database.Open(database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer func() {
		log.WithFields(logrus.Fields{
			"status": "stopping database support",
			"host":   cfg.DB.Host,
		}).Info("shutdown")
		db.Close()
	}()
	// ============================================================
	// Start Debug Service
	log.WithFields(logrus.Fields{
		"status": "debug router started",
		"host":   cfg.Web.DebugHost,
	}).Info("startup")

	// The Debug function return a mux to listen and serve on for all the debug
	// related endpoints. This include the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log, db)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.

	// listing on port 4000
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.WithFields(logrus.Fields{
				"status": "debug router closed",
				"host":   cfg.Web.DebugHost,
				"ERROR":  err,
			}).Error("shutdown")
		}
	}()

	// ============================================================
	// Start API Service

	log.WithFields(logrus.Fields{
		"status": "initializing API support",
	}).Info("startup")

	// buffered channel of 1 means sends happens before receive.
	shutdown := make(chan os.Signal, 1)
	// SIGINT - ctl +c
	// SIGTERM - k8s shutdown signal
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// APIMuxConfig contains all the mandatory systems required by handlers.
	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
		DB:       db,
	})

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}
	// Make a channel to listen for errors coming from the listener.
	// Use a buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)
	// Start the service listing for api requests.
	// listing on port 3000, when the request comes in it is given to the mux
	// who decides what handler to use.
	// this goroutine creates other child goroutines.
	// every request executes on it's own goroutine.
	// the goroutines needs to terminate before the parent goroutine.

	go func() {
		log.WithFields(logrus.Fields{
			"status": "api router started",
			"host":   api.Addr,
		}).Info("startup")
		serverErrors <- api.ListenAndServe()
	}()

	// ============================================================
	// Shutdown

	// Blocking main and wait for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.WithFields(logrus.Fields{
			"status": "shutdown started",
			"signal": sig,
		}).Info("shutdown")

		defer log.WithFields(logrus.Fields{
			"status": "shutdown complete",
			"signal": sig,
		}).Info("shutdown")
		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutDownTime)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}
	return nil
}
