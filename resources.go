package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
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
	// skip past first token
	_, _ = jd.Token()
	stack, err := buildObjectXMLStack(jd, el)
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
