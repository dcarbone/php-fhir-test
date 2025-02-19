package main

import (
	"golang.org/x/mod/semver"
)

type FHIRVersion string

const (
	fhirVersionUnknown FHIRVersion = ""

	fhirVersionDSTU1 FHIRVersion = "DSTU1"
	fhirVersionDSTU2 FHIRVersion = "DSTU2"
	fhirVersionSTU3  FHIRVersion = "STU3"
	fhirVersionR4    FHIRVersion = "R4"
	fhirVersionR4B   FHIRVersion = "R4B"
	fhirVersionR5    FHIRVersion = "R5"

	fhirVersionMock FHIRVersion = "mock"
)

func (fv FHIRVersion) SemanticVersion() string {
	switch fv {
	case fhirVersionUnknown:
		return ""
	case fhirVersionMock:
		return "v99.99.99"

	case fhirVersionDSTU1:
		return "v0.0.82"
	case fhirVersionDSTU2:
		return "v1.0.2"
	case fhirVersionSTU3:
		return "v3.0.1"
	case fhirVersionR4:
		return "v4.0.0"
	case fhirVersionR4B:
		return "v4.3.0"
	case fhirVersionR5:
		return "v5.0.0"

	default:
		// todo: should this cause an error?
		return "v99.99.99"
	}
}

func (fv FHIRVersion) ShortVersion() string {
	switch fv {
	case fhirVersionUnknown:
		return ""
	case fhirVersionMock:
		return "v99.99"

	case fhirVersionDSTU1:
		return "v0.0"
	case fhirVersionDSTU2:
		return "v1.0"
	case fhirVersionSTU3:
		return "v3.0"
	case fhirVersionR4:
		return "v4.0"
	case fhirVersionR4B:
		return "v4.3"
	case fhirVersionR5:
		return "v5.0"

	default:
		return "v99.99"
	}
}

func fhirVersionFromString(fv string) FHIRVersion {
	switch true {
	case semver.Compare(fv, "v1.0.0") == -1:
		return fhirVersionDSTU1
	case semver.Compare(fv, "v1.0.0") >= 0 && semver.Compare(fv, "v3.0.0") == -1:
		return fhirVersionDSTU2
	case semver.Compare(fv, "v3.0.0") >= 0 && semver.Compare(fv, "v4.0.0") == -1:
		return fhirVersionSTU3
	case semver.Compare(fv, "v4.0.0") >= 0 && semver.Compare(fv, "v4.3.0") == -1:
		return fhirVersionR4
	case semver.Compare(fv, "v4.3.0") >= 0 && semver.Compare(fv, "v5.0.0") == -1:
		return fhirVersionR4B
	case semver.Compare(fv, "v5.0.0") >= 0 && semver.Compare(fv, "v6.0.0") == -1:
		return fhirVersionR5

	default:
		return fhirVersionMock
	}
}
