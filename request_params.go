package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type requestParamsCtxKeyT struct{}

var requestParamsCtxKey requestParamsCtxKeyT

type RequestParams struct {
	AcceptFormat  SerializeFormat
	AcceptVersion FHIRVersion

	ContentTypeFormat  SerializeFormat
	ContentTypeVersion FHIRVersion

	Count int
}

func getRequestParams(r *http.Request) RequestParams {
	rp := r.Context().Value(requestParamsCtxKey).(*RequestParams)
	return *rp
}

func extractFormatAndVersionFromHeader(hdrs []string) (SerializeFormat, FHIRVersion) {
hdrLoop:
	for _, hdr := range hdrs {
		var (
			ctyp string
			vstr string
		)
		vs := strings.SplitN(hdr, ";", 2)
		switch len(vs) {
		case 2:
			ctyp, vstr = vs[0], vs[1]
		case 1:
			ctyp = vs[0]

		default:
			continue hdrLoop
		}

		ctyp = strings.ToLower(strings.TrimSpace(ctyp))
		vstr = strings.ToLower(strings.TrimSpace(vstr))

		// skip past anything that isn't formatted as "application/$format"
		if !strings.HasPrefix(ctyp, "application/") {
			continue hdrLoop
		}

		// handle "Accept: application/fhir+json, application/json+fhir" sort of stuff.
		// todo: for now, only bother with the first seen value
		if strings.Contains(ctyp, ",") {
			ctyp = strings.TrimSpace(strings.SplitN(ctyp, ",", 2)[0])
		}

		// strip prefix
		ctyp = strings.TrimPrefix(ctyp, "application/")

		// handle "application/json" and "application/xml"
		if ctyp == "xml" || ctyp == "json" {
			ctyp = fmt.Sprintf("fhir+%s", ctyp)
		}

		// check if remaining value is valid.
		format := SerializeFormat(ctyp)
		if !format.Valid() {
			continue hdrLoop
		}

		// if version is empty, move on.
		if vstr == "" {
			return format, fhirVersionUnknown
		}

		// otherwise, set fhir version and move on.
		return format, fhirVersionFromString(vstr)
	}

	return "", fhirVersionUnknown
}

func parseQueryFormatParam(r *http.Request) SerializeFormat {
	return SerializeFormat(strings.ToLower(r.URL.Query().Get("_format")))
}

func middlewareParseRequestParams(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			rp                 RequestParams
			acceptFormat       SerializeFormat
			acceptVersion      FHIRVersion
			contentTypeFormat  SerializeFormat
			contentTypeVersion FHIRVersion
			err                error
		)

		log := getRequestLogger(r)

		// first, require _format param to either be empty or valid.
		queryFormat := parseQueryFormatParam(r)
		if queryFormat != SerializeFormatUnknown && !queryFormat.Valid() {
			log.Error("Invalid _format query param value", "_format", string(queryFormat))
			http.Error(w, fmt.Sprintf(`_format query param value %q not in [xml json]`, string(queryFormat)), http.StatusBadRequest)
			return
		}

		// extract accept format
		acceptFormat, acceptVersion = extractFormatAndVersionFromHeader(r.Header.Values("accept"))

		// if the query param is set, accept and content type headers must agree.
		if queryFormat != SerializeFormatUnknown {
			if acceptFormat == SerializeFormatUnknown {
				acceptFormat = queryFormat
			} else if (queryFormat.IsJson() && !acceptFormat.IsJson()) || (queryFormat.IsXml() && !acceptFormat.IsXml()) {
				log.Error("Query param _format disagrees with Accept header", "_format", "json", "Accept", string(acceptFormat))
				http.Error(w, fmt.Sprintf("query param _format value %q disagrees with Accept value %q", string(queryFormat), string(acceptFormat)), http.StatusBadRequest)
				return
			}
		}

		// if dealing with a request tha contains a body, try to extract content type header details
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			contentTypeFormat, contentTypeVersion = extractFormatAndVersionFromHeader(r.Header.Values("content-type"))

			if queryFormat != SerializeFormatUnknown {
				if contentTypeFormat == SerializeFormatUnknown {
					contentTypeFormat = queryFormat
				} else if (queryFormat.IsJson() && !contentTypeFormat.IsJson()) || (queryFormat.IsXml() && !contentTypeFormat.IsXml()) {
					log.Error("Query param _format disagrees with Content-Type header", "_format", "json", "Accept", string(contentTypeFormat))
					http.Error(w, fmt.Sprintf("query param _format value %q disagrees with Accept value %q", string(queryFormat), string(contentTypeFormat)), http.StatusBadRequest)
					return
				}
			}
		}

		// parse _count param, if found
		countstr := r.URL.Query().Get("_count")
		if countstr != "" {
			if rp.Count, err = strconv.Atoi(countstr); err != nil {
				log.Error("Cannot parse _count query param", "_count", countstr, "err", err)
				http.Error(w, fmt.Sprintf("_count query param value %q not parseable as int: %v", countstr, err), http.StatusBadRequest)
				return
			} else if rp.Count < 0 {
				log.Error("Negative _count query param value seen", "_count", rp.Count)
				http.Error(w, "_count query param value must be >= 0", http.StatusBadRequest)
				return
			}
		}

		// set values
		rp.AcceptFormat = acceptFormat
		rp.AcceptVersion = acceptVersion
		rp.ContentTypeFormat = contentTypeFormat
		rp.ContentTypeVersion = contentTypeVersion

		// call next handler, embedding the parsed params into the request context.
		next(w, r.WithContext(context.WithValue(r.Context(), requestParamsCtxKey, &rp)))
	}
}
