package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

func buildResponseContentTypeHeader(rp RequestParams) string {
	ctype := fmt.Sprintf("application/%s", string(rp.AcceptFormat))
	if rp.AcceptVersion != FHIRVersionUnknown {
		ctype = fmt.Sprintf("%s; fhirVersion=%s", ctype, rp.AcceptVersion.ShortVersion())
	}
	return ctype
}

func respondInKind(w http.ResponseWriter, r *http.Request, data any) {
	var (
		err error

		log = getRequestLogger(r)
		rp  = getRequestParams(r)
	)

	w.Header().Set("Content-Type", buildResponseContentTypeHeader(rp))

	switch true {
	case rp.AcceptFormat.IsJson():
		if err = json.NewEncoder(w).Encode(data); err != nil {
			log.Error("Error during JSON encode", "data", fmt.Sprintf("%T", data), "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case rp.AcceptFormat.IsXml():
		// write header
		if _, err = w.Write([]byte(xml.Header)); err != nil {
			log.Error("Error writing XML lead in", "err", err)
			http.Error(w, "error writing XML lead in", http.StatusInternalServerError)
			return
		}

		// init xml encoder
		xe := xml.NewEncoder(w)
		defer func() { _ = xe.Close() }()

		if err = xe.Encode(data); err != nil {
			log.Error("Error during XML encode", "data", fmt.Sprintf("%T", data), "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	default:
		log.Error("Unknown format specified", "format", string(rp.AcceptFormat))
		http.Error(w, "unknown format specified", http.StatusBadRequest)
	}
}
