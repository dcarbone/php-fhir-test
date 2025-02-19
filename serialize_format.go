package main

import (
	"slices"
)

type SerializeFormat string

const (
	SerializeFormatUnknown    SerializeFormat = ""
	SerializeFormatXml        SerializeFormat = "fhir+xml"
	SerializeFormatJson       SerializeFormat = "fhir+json"
	SerializeFormatXmlLegacy  SerializeFormat = "xml+fhir"
	SerializeFormatJsonLegacy SerializeFormat = "json+fhir"
	SerializeFormatXmlPatch   SerializeFormat = "xml-patch+xml"
	SerializeFormatJsonPatch  SerializeFormat = "json-patch+json"
)

var (
	serializeFormats = []SerializeFormat{
		SerializeFormatXml,
		SerializeFormatJson,
		SerializeFormatXmlLegacy,
		SerializeFormatJsonLegacy,
		SerializeFormatXmlPatch,
		SerializeFormatJsonPatch,
	}
)

func (sf SerializeFormat) Valid() bool {
	return slices.Contains(serializeFormats, sf)
}

func (sf SerializeFormat) IsLegacy() bool {
	return SerializeFormatXmlPatch == sf || SerializeFormatJsonLegacy == sf
}

func (sf SerializeFormat) IsPatch() bool {
	return SerializeFormatXmlPatch == sf || SerializeFormatJsonPatch == sf
}

func (sf SerializeFormat) IsXml() bool {
	return SerializeFormatXml == sf || SerializeFormatXmlLegacy == sf
}

func (sf SerializeFormat) IsJson() bool {
	return SerializeFormatJson == sf || SerializeFormatJsonLegacy == sf
}

func serializeFormatStrings() []string {
	out := make([]string, len(serializeFormats))
	for i := range serializeFormats {
		out[i] = string(serializeFormats[i])
	}
	return out
}
