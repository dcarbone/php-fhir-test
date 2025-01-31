package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
)

var (
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
	resourceMap = make(map[string]map[string][]*Resource)
)

type Resource struct {
	ResourceType string
	ID           string
	Data         []byte
	XMLNS        bool
}

func (r *Resource) UnmarshalJSON(b []byte) error {
	type miniRes struct {
		ResourceType string `json:"resourceType"`
		ID           string `json:"id"`
	}
	tmp := new(miniRes)
	if err := json.Unmarshal(b, &tmp); err != nil {
		return fmt.Errorf("error unmarshalling resource: %w", err)
	}
	r.ResourceType = tmp.ResourceType
	r.ID = tmp.ID
	r.Data = make([]byte, len(b))
	copy(r.Data, b)
	return nil
}

func (r Resource) MarshalJSON() ([]byte, error) {
	return r.Data, nil
}

func (r Resource) MarshalXML(xe *xml.Encoder, _ xml.StartElement) error {
	el := &xml.StartElement{
		Name: xml.Name{Local: r.ResourceType},
	}
	if r.XMLNS {
		el.Attr = []xml.Attr{
			{
				Name:  xml.Name{Local: "xmlns"},
				Value: "https://hl7.org/fhir",
			},
		}
	}

	jd := json.NewDecoder(bytes.NewReader(r.Data))
	jd.UseNumber()

	// skip past first token
	_, _ = jd.Token()
	stack, err := buildObjectXMLStack(r, jd, el)
	if err != nil {
		return err
	}

	stack = append([]any{el}, stack...)
	stack = append(stack, el.End())

	return encodeXMLStack(xe, stack)
}

type BundleEntry struct {
	Resource *Resource `json:"resource"`
}

func (be BundleEntry) MarshalXML(xe *xml.Encoder, _ xml.StartElement) error {
	var err error
	el := xml.StartElement{Name: xml.Name{Local: "resource"}}
	if err = xe.EncodeToken(el); err != nil {
		return err
	}
	if err = xe.Encode(be.Resource); err != nil {
		return err
	}
	return xe.EncodeToken(el.End())
}

type Bundle struct {
	ResourceType string        `json:"resourceType" xml:"-"`
	Entry        []BundleEntry `json:"entry" xml:"entry"`
}

func (b Bundle) MarshalXML(xe *xml.Encoder, _ xml.StartElement) error {
	var err error

	el := xml.StartElement{
		Name: xml.Name{Local: "Bundle"},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "xmlns"},
				Value: "https://hl7.org/fhir",
			},
		},
	}

	if err = xe.EncodeToken(el); err != nil {
		return err
	}

	for _, e := range b.Entry {
		en := xml.StartElement{Name: xml.Name{Local: "entry"}}
		if err = xe.EncodeToken(en); err != nil {
			return err
		}
		if err = xe.Encode(e); err != nil {
			return err
		}
		if err = xe.EncodeToken(en.End()); err != nil {
			return err
		}
	}

	return xe.EncodeToken(el.End())
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
			resourceMap[fv][res.ResourceType] = make([]*Resource, 0)
		}
		resourceMap[fv][res.ResourceType] = append(resourceMap[fv][res.ResourceType], res)
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

		name := strings.TrimPrefix(th.Name, "./")
		if name == "" {
			continue
		}

		switch th.Typeflag {
		case tar.TypeDir:
			log.Info("Found directory", "dir", name)
			fv = strings.ToUpper(filepath.Base(name))
			if _, ok := resourceMap[fv]; !ok {
				resourceMap[fv] = make(map[string][]*Resource)
			}
		case tar.TypeReg:
			if err = parseResources(ctx, tr, th, fv); err != nil {
				return fmt.Errorf("error parsing resources from file %q in version %q: %w", name, fv, err)
			}
		default:
			log.Warn("Found unexpected file type", "type", string(th.Typeflag))
		}
	}
}
