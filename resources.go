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
	"slices"
	"sort"
	"strings"
	"sync"
)

var (
	versionResourceMapMu sync.RWMutex
	versionResourceMap   map[FHIRVersion]*ResourceMap
)

func init() {
	versionResourceMap = make(map[FHIRVersion]*ResourceMap)
	for fv := FHIRVersionDSTU1; fv <= FHIRVersionMock; fv++ {
		versionResourceMap[fv] = newResourceMap(fv)
	}
}

type Resource struct {
	FHIRVersion  FHIRVersion `json:"-"`
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
	stack, _, err := buildObjectXMLStack(r, jd, el)
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

type ResourceMap struct {
	mu sync.RWMutex

	version   FHIRVersion
	resources map[string][]*Resource
}

func newResourceMap(fv FHIRVersion) *ResourceMap {
	rm := &ResourceMap{
		version:   fv,
		resources: make(map[string][]*Resource),
	}

	return rm
}

func (rm *ResourceMap) Version() FHIRVersion {
	return rm.version
}

func (rm *ResourceMap) resourceTypes() []string {
	out := make([]string, len(rm.resources))
	i := 0
	for n := range rm.resources {
		out[i] = n
		i++
	}
	sort.Strings(out)
	return out
}

func (rm *ResourceMap) ResourceTypes() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.resourceTypes()
}

func (rm *ResourceMap) GetResourcesByType(resType string, count int) []*Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	rscs := rm.resources[resType]
	if l := len(rscs); count > l || count == 0 {
		count = l
	}
	out := make([]*Resource, count)
	copy(out, rscs)
	return out
}

func (rm *ResourceMap) GetResource(rscType, rscId string) *Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	for _, rsc := range rm.resources[rscType] {
		if rsc.ID == rscId {
			return rsc
		}
	}
	return nil
}

func (rm *ResourceMap) PutResource(rsc *Resource) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// check if this resource already exists in this map, overriding if so.
	for i := range rm.resources[rsc.ResourceType] {
		if rm.resources[rsc.ResourceType][i].ID == rsc.ID {
			rm.resources[rsc.ResourceType][i] = rsc
			return
		}
	}

	// if we make it down here, add as new resource.
	rm.resources[rsc.ResourceType] = append(rm.resources[rsc.ResourceType], rsc)
}

func (rm *ResourceMap) MarshalJSON() ([]byte, error) {
	type rscMap struct {
		FHIRVersion FHIRVersion `json:"fhirVersion"`
		Resources   []string    `json:"resources"`
	}
	out := rscMap{
		FHIRVersion: rm.version,
		Resources:   rm.ResourceTypes(),
	}
	return json.Marshal(out)
}

func (rm *ResourceMap) MarshalXML(xe *xml.Encoder, _ xml.StartElement) error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	fvel := xml.StartElement{
		Name: xml.Name{Local: "FHIRVersion"},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "value"},
				Value: rm.version.String(),
			},
		},
	}
	if err := xe.EncodeToken(fvel); err != nil {
		return err
	}
	rscel := xml.StartElement{Name: xml.Name{Local: "Resources"}}
	if err := xe.EncodeToken(rscel); err != nil {
		return err
	}
	for _, n := range rm.resourceTypes() {
		el := xml.StartElement{
			Name: xml.Name{Local: "Resource"},
			Attr: []xml.Attr{
				{
					Name:  xml.Name{Local: "value"},
					Value: n,
				},
			},
		}
		if err := xe.EncodeElement("", el); err != nil {
			return err
		}
	}
	return errors.Join(
		xe.EncodeToken(rscel.End()),
		xe.EncodeToken(fvel.End()),
	)
}

func versionList() FHIRVersions {
	out := make(FHIRVersions, len(versionResourceMap))
	i := 0
	for fv := range versionResourceMap {
		out[i] = fv
		i++
	}
	slices.SortFunc(out, fhirVersionSemanticSortFunc(true))
	return out
}

func parseSeedResources(ctx context.Context, tr *tar.Reader, th *tar.Header, fv FHIRVersion) error {
	dec := json.NewDecoder(tr)

	i := 0
	for dec.More() {
		if err := ctx.Err(); err != nil {
			return err
		}
		rsc := new(Resource)
		if err := dec.Decode(rsc); err != nil {
			return fmt.Errorf("error decoding resource %d in file %q: %w", i, th.Name, err)
		}
		if rsc.ResourceType == "" {
			return fmt.Errorf("resource %d in file %q has no resourceType value", i, th.Name)
		}
		rsc.FHIRVersion = fv
		// this will panic if you did it wrong.
		versionResourceMap[fv].PutResource(rsc)
	}
	return nil
}

func extractSeedResources(ctx context.Context, log *slog.Logger) error {
	versionResourceMapMu.Lock()
	defer versionResourceMapMu.Unlock()

	var (
		fv FHIRVersion
	)

	log.Info("Seeding FHIR resources from embedded tarball...")

	gr, err := gzip.NewReader(bytes.NewReader(seedResourcesTarball))
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
			fv = fhirVersionFromString(filepath.Base(name))
			if fv == FHIRVersionUnknown || fv == FHIRVersionMock {
				panic(fmt.Sprintf("Cannot seed resources from unnknown FHIR version %q", filepath.Base(name)))
			}
		case tar.TypeReg:
			if err = parseSeedResources(ctx, tr, th, fv); err != nil {
				return fmt.Errorf("error parsing resources from file %q in version %q: %w", name, fv, err)
			}
		default:
			log.Warn("Found unexpected file type", "type", string(th.Typeflag))
		}
	}
}
