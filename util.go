package main

import (
	"path/filepath"
	"slices"
	"strings"
)

func extractFHIRVersionFromDir(in string) string {
	return strings.ToUpper(filepath.Base(in)[5:])
}

func versionList() []string {
	out := make([]string, len(resourceMap))
	i := 0
	for v := range resourceMap {
		out[i] = v
		i++
	}
	slices.Sort(out)
	return out
}

func versionResourceList(fv string) []string {
	out := make([]string, 0)
	for v := range resourceMap[fv] {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}
