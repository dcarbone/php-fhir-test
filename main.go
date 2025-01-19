package main

import (
	"archive/tar"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	//go:embed resources.tar.gz
	resourcesTar []byte

	bindAddr = "127.0.0.1:8080"

	fg = flag.NewFlagSet("fhir-test-server", flag.ContinueOnError)

	/*
		{
			"resourceType": {
				"resourceID": {
					...
				},
				...
			}
		}
	*/
	resourceMap = make(map[string]map[string]Resource)
)

type Resource struct {
	Type   string         `json:"resourceType"`
	ID     string         `json:"-"`
	Fields map[string]any `json:"-"`
}

func (r *Resource) UnmarshalJSON(b []byte) error {
	tmp := make(map[string]any)
	if err := json.Unmarshal(b, &tmp); err != nil {
		return fmt.Errorf("error unmarshalling resource: %w", err)
	}
	r.Type = tmp["resourceType"].(string)
	r.ID = tmp["id"].(string)
	delete(tmp, "resourceType")
	r.Fields = tmp
	return nil
}

func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Fields)
}

type BundleEntry struct {
	FullURL  string   `json:"fullUrl"`
	Resource Resource `json:"resource"`
}

type Bundle struct {
	Version      string        `json:"-"`
	ResourceType string        `json:"resourceType"`
	Type         string        `json:"type"`
	Entry        []BundleEntry `json:"entry"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))

	fg.StringVar(&bindAddr, "bind", bindAddr, "Address, including port, to bind.")
	if err := fg.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Error("Error parsing flags", "err", err)
		os.Exit(1)
	}

	if err := extractResources(ctx, log); err != nil {
		log.Error("Error extracting resources", "err", err)
		os.Exit(1)
	}
}

func extractResources(ctx context.Context, log *slog.Logger) error {
	log.Info("Extracting FHIR resources...")
	tr := tar.NewReader(bytes.NewReader(resourcesTar))

	n, err := tr.Next()
	for err != nil {

	}
}

func parseResources(ctx context.Context) {

}
