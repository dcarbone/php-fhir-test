package main

import (
	"path/filepath"
	"strings"
)

func extractFHIRVersionFromDir(in string) string {
	return strings.ToUpper(filepath.Base(in)[5:])
}
