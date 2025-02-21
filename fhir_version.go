package main

import (
	"encoding/json"
	"encoding/xml"
	"strings"

	"golang.org/x/mod/semver"
)

type FHIRVersion int

const (
	FHIRVersionUnknown FHIRVersion = iota

	FHIRVersionDSTU1
	FHIRVersionDSTU2
	FHIRVersionSTU3
	FHIRVersionR4
	FHIRVersionR4B
	FHIRVersionR5

	// FHIRVersionMock is used as a fallback for all non-empty versions.  It must always have the highest enum value.
	FHIRVersionMock
)

func (fv FHIRVersion) String() string {
	switch fv {
	case FHIRVersionUnknown:
		return ""

	case FHIRVersionDSTU1:
		return "DSTU1"
	case FHIRVersionDSTU2:
		return "DSTU2"
	case FHIRVersionSTU3:
		return "STU3"
	case FHIRVersionR4:
		return "R4"
	case FHIRVersionR4B:
		return "R4B"
	case FHIRVersionR5:
		return "R5"

	default:
		return "mock"
	}
}

func (fv FHIRVersion) SemanticVersion() string {
	switch fv {
	case FHIRVersionUnknown:
		return ""

	case FHIRVersionDSTU1:
		return "v0.0.82"
	case FHIRVersionDSTU2:
		return "v1.0.2"
	case FHIRVersionSTU3:
		return "v3.0.1"
	case FHIRVersionR4:
		return "v4.0.0"
	case FHIRVersionR4B:
		return "v4.3.0"
	case FHIRVersionR5:
		return "v5.0.0"

	default:
		// mock
		return "v99.99.99"
	}
}

func (fv FHIRVersion) ShortVersion() string {
	switch fv {
	case FHIRVersionUnknown:
		return ""

	case FHIRVersionDSTU1:
		return "v0.0"
	case FHIRVersionDSTU2:
		return "v1.0"
	case FHIRVersionSTU3:
		return "v3.0"
	case FHIRVersionR4:
		return "v4.0"
	case FHIRVersionR4B:
		return "v4.3"
	case FHIRVersionR5:
		return "v5.0"

	default:
		// mock
		return "v99.99"
	}
}

func (fv FHIRVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(fv.String())
}

func (fv FHIRVersion) MarshalXML(xe *xml.Encoder, _ xml.StartElement) error {
	el := xml.StartElement{
		Name: xml.Name{Local: "FHIRVersion"},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "value"},
				Value: fv.String(),
			},
		},
	}
	return xe.EncodeElement("", el)
}

func fhirVersionFromString(fv string) FHIRVersion {
	fv = strings.ToUpper(fv)
	switch true {
	case fv == "":
		return FHIRVersionUnknown

	case fv == FHIRVersionDSTU1.String() || semver.Compare(fv, "v1.0.0") == 1:
		return FHIRVersionDSTU1
	case fv == FHIRVersionDSTU2.String() || semver.Compare(fv, "v1.0.0") >= 0 && semver.Compare(fv, "v3.0.0") == 1:
		return FHIRVersionDSTU2
	case fv == FHIRVersionSTU3.String() || semver.Compare(fv, "v3.0.0") >= 0 && semver.Compare(fv, "v4.0.0") == 1:
		return FHIRVersionSTU3
	case fv == FHIRVersionR4.String() || semver.Compare(fv, "v4.0.0") >= 0 && semver.Compare(fv, "v4.3.0") == 1:
		return FHIRVersionR4
	case fv == FHIRVersionR4B.String() || semver.Compare(fv, "v4.3.0") >= 0 && semver.Compare(fv, "v5.0.0") == 1:
		return FHIRVersionR4B
	case fv == FHIRVersionR5.String() || semver.Compare(fv, "v5.0.0") >= 0 && semver.Compare(fv, "v6.0.0") == 1:
		return FHIRVersionR5

	default:
		return FHIRVersionMock
	}
}

func fhirVersionNameSortFunc(asc bool) func(a, b FHIRVersion) int {
	return func(a, b FHIRVersion) int {
		d := strings.Compare(a.String(), b.String())
		if d == 0 || asc {
			return d
		} else {
			return -d
		}
	}
}

func fhirVersionSemanticSortFunc(asc bool) func(a, b FHIRVersion) int {
	return func(a, b FHIRVersion) int {
		d := semver.Compare(a.SemanticVersion(), b.SemanticVersion())
		if d == 0 || asc {
			return d
		} else {
			return -d
		}
	}
}

type FHIRVersions []FHIRVersion

func (fvs FHIRVersions) MarshalJSON() ([]byte, error) {
	out := make([]string, len(fvs))
	for i, fv := range fvs {
		out[i] = fv.String()
	}
	return json.Marshal(out)
}

func (fvs FHIRVersions) MarshalXML(xe *xml.Encoder, _ xml.StartElement) error {
	el := xml.StartElement{
		Name: xml.Name{Local: "FHIRVersions"},
	}
	if err := xe.EncodeToken(el); err != nil {
		return err
	}
	for _, fv := range fvs {
		if err := xe.Encode(fv); err != nil {
			return err
		}
	}
	return xe.EncodeToken(el.End())
}
