package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	//go:embed resources.tar.gz
	resourcesTar []byte

	bindAddr = "127.0.0.1:8080"

	fg = flag.NewFlagSet("fhir-test-server", flag.ContinueOnError)
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key != "source" {
					return a
				}
				return slog.Attr{
					Key: a.Key,
					Value: slog.StringValue(
						strings.Replace(filepath.Base(strings.Trim(a.Value.String(), "{}")), " ", ":", 1),
					),
				}
			},
		},
	))

	fg.StringVar(&bindAddr, "bind", bindAddr, "Address, including port, to bind.")
	if err := fg.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Error("Error parsing flags", "err", err)
		os.Exit(1)
	}

	if err := extractSeedResources(ctx, log); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info("Stopping.")
			os.Exit(0)
		}
		log.Error("Error extracting resources", "err", err)
		os.Exit(1)
	}

	ch := make(chan error)

	go func() {
		ch <- runWebserver(log)
	}()

	select {
	case <-ctx.Done():
		log.Info("Shutting down...")
		os.Exit(0)
	case err := <-ch:
		log.Error("Server stopped", "err", err)
		os.Exit(1)
	}
}
