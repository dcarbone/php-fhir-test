package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
)

type (
	loggerCtxKeyT struct{}
)

var (
	errInvalidFormat = errors.New("invalid format")

	requestIdSource atomic.Uint64

	loggerCtxKey loggerCtxKeyT
)

func middlewareEmbedLogger(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := requestIdSource.Add(1)
		log = log.With("rid", rid)
		log.Info("Processing request...", "method", r.Method, "url", r.URL)
		r.WithContext(context.WithValue(r.Context(), loggerCtxKey, log))
		next(w, r)
	}
}

func getRequestLogger(r *http.Request) *slog.Logger {
	return r.Context().Value(loggerCtxKey).(*slog.Logger)
}

func respondInKind(log *slog.Logger, rp RequestParams, w http.ResponseWriter, fv string, data any) {
	var (
		contentTypeFmt string
		err            error
	)

	switch fv {
	case "DSTU1", "R1", "DSTU2", "R2":
		contentTypeFmt = "%s+fhir"
	case "STU3", "R3", "R4", "R4B", "R5":
		contentTypeFmt = "fhir+%s"
	default:
		contentTypeFmt = "fhir+%s"
	}

	switch rp.Format {
	case "", "json":
		w.Header().Set("Content-Type", fmt.Sprintf("application/"+contentTypeFmt, "json"))
		if err = json.NewEncoder(w).Encode(data); err != nil {
			log.Error("Error during JSON encode", "data", fmt.Sprintf("%T", data), "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "xml":
		w.Header().Set("Content-Type", fmt.Sprintf("application/"+contentTypeFmt, "xml"))

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
		log.Error("Unknown format specified", "format", rp.Format)
		http.Error(w, "unknown format specified", http.StatusBadRequest)
	}
}

func handlerGetVersionList(log *slog.Logger) http.HandlerFunc {
	return middlewareEmbedLogger(log, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" && r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(versionList()); err != nil {
			log.Error("Error encoding version list", "err", err)
			http.Error(w, "Error encoding version list", http.StatusInternalServerError)
		}
	})
}

func handlerGetVersionResourceList(log *slog.Logger, fv string) http.HandlerFunc {
	return middlewareEmbedLogger(log, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf("/%s", fv) && r.URL.Path != fmt.Sprintf("/%s/", fv) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(versionResourceList(fv)); err != nil {
			log.Error("Error encoding version resource list", "version", fv, "err", err)
			http.Error(w, fmt.Sprintf("Error encoding version %q resource list", fv), http.StatusInternalServerError)
		}
	})
}

func handlerGetResourceBundle(fv FHIRVersion, resType string) http.HandlerFunc {
	return middlewareParseRequestParams(false, func(w http.ResponseWriter, r *http.Request) {
		rp := getRequestParams(r)
		cnt := rp.Count
		if l := len(resourceMap[fv][resType]); cnt == 0 || cnt > l {
			cnt = l
		}

		out := Bundle{
			ResourceType: "Bundle",
			Entry:        make([]BundleEntry, cnt),
		}
		for i := 0; i < cnt; i++ {
			out.Entry[i] = BundleEntry{Resource: resourceMap[fv][resType][i]}
		}

		respondInKind(rp, w, fv, out)
	})
}

func handlerGetVersionResource(log *slog.Logger, fv, resType string, i int) http.HandlerFunc {
	return middlewareEmbedLogger(log, func(w http.ResponseWriter, r *http.Request) {
		rp, err := middlewareParseRequestParams(r)
		if err != nil {
			log.Error("Error parsing query params", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if rp.Count != 0 {
			http.Error(w, "_count must be zero or undefined with specific resource ID", http.StatusBadRequest)
			return
		}

		respondInKind(log, rp, w, fv, resourceMap[fv][resType][i])
	})
}

func runWebserver(log *slog.Logger) error {
	log.Info("Building routes...")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", middlewareEmbedLogger(log, handlerGetVersionList()))

	for fv, resourceTypes := range resourceMap {
		mux.HandleFunc(fmt.Sprintf("GET /%s", fv), middlewareEmbedLogger(log, handlerGetVersionResourceList(fv)))
		mux.HandleFunc(fmt.Sprintf("GET /%s/", fv), middlewareEmbedLogger(log, handlerGetVersionResourceList(fv)))

		for resType, resources := range resourceTypes {
			mux.HandleFunc(fmt.Sprintf("GET /%s/%s", fv, resType), middlewareEmbedLogger(log, handlerGetResourceBundle(fv, resType)))
			mux.HandleFunc(fmt.Sprintf("GET /%s/%s/", fv, resType), middlewareEmbedLogger(log, handlerGetResourceBundle(fv, resType)))

			for i, res := range resources {
				if res.ID == "" {
					continue
				}
				mux.HandleFunc(fmt.Sprintf("GET /%s/%s/%s", fv, resType, res.ID), middlewareEmbedLogger(log, handlerGetVersionResource(fv, resType, i)))
				mux.HandleFunc(fmt.Sprintf("GET /%s/%s/%s/", fv, resType, res.ID), middlewareEmbedLogger(log, handlerGetVersionResource(fv, resType, i)))
			}
		}
	}

	log.Info("Webserver running", "addr", bindAddr)

	return http.ListenAndServe(bindAddr, mux)
}
