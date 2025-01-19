package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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
			"fhirVersion" {
				"resourceType": [
					{
						...
					},
					...
				],
				...
			}
		}
	*/
	resourceMap = make(map[string]map[string][]Resource)
)

type Resource struct {
	ResourceType string         `json:"resourceType"`
	ID           string         `json:"-"`
	Fields       map[string]any `json:"-"`
}

func (r *Resource) UnmarshalJSON(b []byte) error {
	tmp := make(map[string]any)
	if err := json.Unmarshal(b, &tmp); err != nil {
		return fmt.Errorf("error unmarshalling resource: %w", err)
	}
	r.ResourceType, _ = tmp["resourceType"].(string)
	r.ID, _ = tmp["id"].(string)
	delete(tmp, "resourceType")
	r.Fields = tmp
	return nil
}

func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Fields)
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
		if errors.Is(err, context.Canceled) {
			log.Info("Stopping.")
			os.Exit(0)
		}
		log.Error("Error extracting resources", "err", err)
		os.Exit(1)
	}
}

func parseResources(ctx context.Context, tr *tar.Reader, th *tar.Header, fv string) error {
	dec := json.NewDecoder(tr)

	i := 0
	for dec.More() {
		if err := ctx.Err(); err != nil {
			return err
		}
		res := new(Resource)
		if err := dec.Decode(res); err != nil {
			return fmt.Errorf("error decoding resource %d in file %q: %w", i, th.Name, err)
		}
		if res.ResourceType == "" {
			return fmt.Errorf("resource %d in file %q has no resourceType value", i, th.Name)
		}
		if _, ok := resourceMap[fv][res.ResourceType]; !ok {
			resourceMap[fv][res.ResourceType] = make([]Resource, 0)
		}
		resourceMap[fv][res.ResourceType] = append(resourceMap[fv][res.ResourceType], *res)
	}
	return nil
}

func extractResources(ctx context.Context, log *slog.Logger) error {
	var (
		fv string
	)

	log.Info("Extracting FHIR resources...")

	gr, err := gzip.NewReader(bytes.NewReader(resourcesTar))
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer func() { _ = gr.Close() }()

	tr := tar.NewReader(gr)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		th, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("error reading tar archive: %w", err)
		}

		if th.Name == "./" {
			continue
		}

		switch th.Typeflag {
		case tar.TypeDir:
			log.Info("Found directory", "dir", th.Name)
			fv = extractFHIRVersionFromDir(th.Name)
			if _, ok := resourceMap[fv]; !ok {
				resourceMap[fv] = make(map[string][]Resource)
			}
		case tar.TypeReg:
			if err = parseResources(ctx, tr, th, fv); err != nil {
				return fmt.Errorf("error parsing resources from file %q in version %q: %w", th.Name, fv, err)
			}
		default:
			log.Warn("Found unexpected file type", "type", string(th.Typeflag))
		}
	}
}
