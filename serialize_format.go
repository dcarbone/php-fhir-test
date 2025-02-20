package main

import (
	"slices"
	"strings"
)

type SerializeFormat string

const (
	SerializeFormatUnknown        SerializeFormat = ""
	SerializeFormatXml            SerializeFormat = "xml"
	SerializeFormatJson           SerializeFormat = "json"
	SerializeFormatFhirXml        SerializeFormat = "fhir+xml"
	SerializeFormatFhirJson       SerializeFormat = "fhir+json"
	SerializeFormatFhirXmlLegacy  SerializeFormat = "xml+fhir"
	SerializeFormatFhirJsonLegacy SerializeFormat = "json+fhir"
	SerializeFormatFhirXmlPatch   SerializeFormat = "xml-patch+xml"
	SerializeFormatFhirJsonPatch  SerializeFormat = "json-patch+json"
)

var (
	serializeFormats = []SerializeFormat{
		SerializeFormatXml,
		SerializeFormatJson,
		SerializeFormatFhirXml,
		SerializeFormatFhirJson,
		SerializeFormatFhirXmlLegacy,
		SerializeFormatFhirJsonLegacy,
		SerializeFormatFhirXmlPatch,
		SerializeFormatFhirJsonPatch,
	}
)

func (sf SerializeFormat) Valid() bool {
	return slices.Contains(serializeFormats, sf)
}

func (sf SerializeFormat) IsFHIR() bool {
	return strings.Contains(string(sf), "fhir") ||
		sf == SerializeFormatFhirXmlPatch ||
		sf == SerializeFormatFhirJsonPatch
}

func (sf SerializeFormat) IsFHIRLegacy() bool {
	return sf == SerializeFormatFhirXmlPatch || sf == SerializeFormatFhirJsonLegacy
}

func (sf SerializeFormat) IsFHIRPatch() bool {
	return sf == SerializeFormatFhirXmlPatch || sf == SerializeFormatFhirJsonPatch
}

func (sf SerializeFormat) IsXml() bool {
	return sf == SerializeFormatXml || sf == SerializeFormatFhirXml || sf == SerializeFormatFhirXmlLegacy
}

func (sf SerializeFormat) IsJson() bool {
	return sf == SerializeFormatJson || sf == SerializeFormatFhirJson || sf == SerializeFormatFhirJsonLegacy
}

func serializeFormatStrings() []string {
	out := make([]string, len(serializeFormats))
	for i := range serializeFormats {
		out[i] = string(serializeFormats[i])
	}
	return out
}
