package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/lmittmann/tint"
	"github.com/spf13/pflag"
	"libdb.so/comd"
	"libdb.so/hserve"
)

var (
	listenAddr = ":8080"
	configFile = "comd.example.json"
	logFormat  = "color"
	verbose    = false
	silent     = false
)

func init() {
	pflag.StringVarP(&listenAddr, "listen-addr", "l", listenAddr, "HTTP address to listen on")
	pflag.StringVarP(&configFile, "config-file", "c", configFile, "path to the configuration file")
	pflag.StringVar(&logFormat, "log-format", logFormat, "log format (color, text, json)")
	pflag.BoolVarP(&silent, "silent", "s", silent, "suppress all output except errors")
	pflag.BoolVarP(&verbose, "verbose", "v", verbose, "increase verbosity level, overrides --silent (info by default, -v for debug)")
	pflag.Usage = func() {
		o := os.Stderr
		fmt.Fprintf(o, "About:\n")
		fmt.Fprintf(o, "  comd implements a Command Daemon that listens for HTTP requests to execute commands.\n")
		fmt.Fprintf(o, "  For more information, see https://libdb.so/comd.\n")
		fmt.Fprintf(o, "\n")
		fmt.Fprintf(o, "Usage:\n")
		fmt.Fprintf(o, "  comd [flags]\n")
		fmt.Fprintf(o, "\n")
		fmt.Fprintf(o, "Flags:\n")
		pflag.PrintDefaults()
	}
}

func main() {
	pflag.Parse()

	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	} else if silent {
		logLevel = slog.LevelWarn
	}

	var logHandler slog.Handler
	switch {
	case logFormat == "json":
		logHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	case logFormat == "color" && os.Getenv("NOCOLOR") == "":
		logHandler = tint.NewHandler(os.Stderr, &tint.Options{Level: logLevel})
	default:
		logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	}
	slog.SetDefault(slog.New(logHandler))

	cfg, err := readConfig()
	if err != nil {
		slog.Error("cannot read config", tint.Err(err))
		os.Exit(1)
	}

	handler := comd.NewHTTPServer(&cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	slog.Info("starting server", "listen_addr", listenAddr)

	if err := hserve.ListenAndServe(ctx, listenAddr, handler); err != nil {
		slog.Error("cannot start server", tint.Err(err))
		os.Exit(1)
	}
}

func readConfig() (comd.Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return comd.Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg comd.Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return comd.Config{}, fmt.Errorf("failed to decode config file as JSON: %w", err)
	}

	return cfg, nil
}
