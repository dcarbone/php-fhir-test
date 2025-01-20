package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
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

func parseRequestParams(r *http.Request) (RequestParams, error) {
	var (
		rp  RequestParams
		err error
	)

	format := r.URL.Query().Get("_format")
	switch format {
	case "json", "xml":
		rp.Format = format
	case "":
		rp.Format = "json"
	default:
		return rp, fmt.Errorf("unknown format: %s", format)
	}

	countstr := r.URL.Query().Get("_count")
	if countstr != "" {
		if rp.Count, err = strconv.Atoi(countstr); err != nil {
			return rp, fmt.Errorf("invalid value provided to _count: %s", countstr)
		}
	}

	return rp, nil
}
