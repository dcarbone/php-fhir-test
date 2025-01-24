package main

import (
	"encoding/json"
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

type BundleEntry struct {
	Resource *Resource `json:"resource"`
}
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	Entry        []BundleEntry `json:"entry"`
}
