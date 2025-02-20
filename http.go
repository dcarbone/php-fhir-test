package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
)

type (
	loggerCtxKeyT struct{}
)

var (
	requestIdSource atomic.Uint64

	loggerCtxKey loggerCtxKeyT
)

func middlewareEmbedLogger(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rid := requestIdSource.Add(1)
		log := log.With("rid", rid)
		log.Info("Processing request...", "method", r.Method, "url", r.URL)
		next(w, r.WithContext(context.WithValue(r.Context(), loggerCtxKey, log)))
	}
}

func getRequestLogger(r *http.Request) *slog.Logger {
	return r.Context().Value(loggerCtxKey).(*slog.Logger)
}

func handlerGetVersionList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		} else {
			respondInKind(w, r, versionList())
		}
	}
}

func handlerGetVersionResourceList(fv FHIRVersion) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf("/%s", fv) && r.URL.Path != fmt.Sprintf("/%s/", fv) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		respondInKind(w, r, versionResourceList(fv))
	}
}

func handlerGetResourceBundle(fv FHIRVersion, rscType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rp := getRequestParams(r)
		cnt := rp.Count
		if l := len(resourceMap[fv][rscType]); cnt == 0 || cnt > l {
			cnt = l
		}

		out := Bundle{
			ResourceType: "Bundle",
			Entry:        make([]BundleEntry, cnt),
		}
		for i := 0; i < cnt; i++ {
			out.Entry[i] = BundleEntry{Resource: resourceMap[fv][rscType][i]}
		}

		respondInKind(w, r, out)
	}
}

func handlerGetVersionResource(fv FHIRVersion, rscType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := getRequestLogger(r)
		rp := getRequestParams(r)
		if rp.Count != 0 {
			http.Error(w, "_count must be zero or undefined with specific resource ID", http.StatusBadRequest)
			return
		}

		resourceId := r.PathValue("resource_id")
		if resourceId == "" {
			log.Error("Unable to parse resource_id param from path")
			http.Error(w, "missing resource_id path parameter", http.StatusBadRequest)
			return
		}

		for _, rsc := range resourceMap[fv][rscType] {
			if rsc.ID == resourceId {
				respondInKind(w, r, rsc)
				return
			}
		}

		log.Error("Resource not found", "resource_id", resourceId)
		http.Error(w, fmt.Sprintf("no version %q resource %q found with id %q", string(fv), rscType, resourceId), http.StatusNotFound)
	}
}

func addHandler(log *slog.Logger, mux *http.ServeMux, route string, hdl http.HandlerFunc) {
	log.Info("Adding route handler", "route", route)

	mux.HandleFunc(route, middlewareEmbedLogger(log.With("route", route), middlewareParseRequestParams(hdl)))
}

func runWebserver(log *slog.Logger) error {
	log.Info("Building routes...")

	mux := http.NewServeMux()

	for fv, resourceTypes := range resourceMap {
		// get version resource list
		addHandler(log, mux, fmt.Sprintf("GET /%s/", string(fv)), handlerGetVersionResourceList(fv))

		for rscType := range resourceTypes {
			// get version resource bundle
			addHandler(log, mux, fmt.Sprintf("GET /%s/%s/", string(fv), rscType), handlerGetResourceBundle(fv, rscType))

			// get specific version resource by id
			addHandler(log, mux, fmt.Sprintf("GET /%s/%s/{resource_id}/", string(fv), rscType), handlerGetVersionResource(fv, rscType))
		}
	}

	// get version list
	addHandler(log, mux, "GET /{$}", handlerGetVersionList())

	log.Info("Webserver running", "addr", bindAddr)

	return http.ListenAndServe(bindAddr, mux)
}
