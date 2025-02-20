package main

import (
	"strings"

	"golang.org/x/mod/semver"
)

type FHIRVersion string

const (
	FHIRVersionUnknown FHIRVersion = ""

	FHIRVersionDSTU1 FHIRVersion = "DSTU1"
	FHIRVersionDSTU2 FHIRVersion = "DSTU2"
	FHIRVersionSTU3  FHIRVersion = "STU3"
	FHIRVersionR4    FHIRVersion = "R4"
	FHIRVersionR4B   FHIRVersion = "R4B"
	FHIRVersionR5    FHIRVersion = "R5"

	FHIRVersionMock FHIRVersion = "mock"
)

func (fv FHIRVersion) String() string {
	return string(fv)
}

func (fv FHIRVersion) MarshalJSON() ([]byte, error) {
	return []byte(fv), nil
}

func (fv FHIRVersion) SemanticVersion() string {
	switch fv {
	case FHIRVersionUnknown:
		return ""
	case FHIRVersionMock:
		return "v99.99.99"

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
		// todo: should this cause an error?
		return "v99.99.99"
	}
}

func (fv FHIRVersion) ShortVersion() string {
	switch fv {
	case FHIRVersionUnknown:
		return ""
	case FHIRVersionMock:
		return "v99.99"

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
		return "v99.99"
	}
}

func fhirVersionFromString(fv string) FHIRVersion {
	fv = strings.ToUpper(fv)
	switch true {
	case fv == string(FHIRVersionDSTU1) || semver.Compare(fv, "v1.0.0") == 1:
		return FHIRVersionDSTU1
	case fv == string(FHIRVersionDSTU2) || semver.Compare(fv, "v1.0.0") >= 0 && semver.Compare(fv, "v3.0.0") == 1:
		return FHIRVersionDSTU2
	case fv == string(FHIRVersionSTU3) || semver.Compare(fv, "v3.0.0") >= 0 && semver.Compare(fv, "v4.0.0") == 1:
		return FHIRVersionSTU3
	case fv == string(FHIRVersionR4) || semver.Compare(fv, "v4.0.0") >= 0 && semver.Compare(fv, "v4.3.0") == 1:
		return FHIRVersionR4
	case fv == string(FHIRVersionR4B) || semver.Compare(fv, "v4.3.0") >= 0 && semver.Compare(fv, "v5.0.0") == 1:
		return FHIRVersionR4B
	case fv == string(FHIRVersionR5) || semver.Compare(fv, "v5.0.0") >= 0 && semver.Compare(fv, "v6.0.0") == 1:
		return FHIRVersionR5

	default:
		return FHIRVersionMock
	}
}

func fhirVersionSortFunc(asc bool) func(a, b FHIRVersion) int {
	return func(a, b FHIRVersion) int {
		d := semver.Compare(a.SemanticVersion(), b.SemanticVersion())
		if d == 0 || asc {
			return d
		} else {
			return -d
		}
	}
}
