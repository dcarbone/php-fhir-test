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
		respondInKind(w, r, versionResourceMap[fv])
	}
}

func handlerGetResourceBundle(fv FHIRVersion, rscType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rp := getRequestParams(r)

		rscs := versionResourceMap[fv].GetResourcesByType(rscType, rp.Count)

		out := Bundle{
			ResourceType: "Bundle",
			Entry:        make([]BundleEntry, len(rscs)),
		}
		for i, rsc := range rscs {
			out.Entry[i] = BundleEntry{Resource: rsc}
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

		rsc := versionResourceMap[fv].GetResource(rscType, resourceId)

		if nil != rsc {
			respondInKind(w, r, rsc)
			return
		}

		log.Error("Resource not found", "resource_id", resourceId)
		http.Error(w, fmt.Sprintf("no version %q resource %q found with id %q", fv.String(), rscType, resourceId), http.StatusNotFound)
	}
}

func addHandler(log *slog.Logger, mux *http.ServeMux, route string, hdl http.HandlerFunc) {
	log.Info("Adding route handler", "route", route)

	mux.HandleFunc(route, middlewareEmbedLogger(log.With("route", route), middlewareParseRequestParams(hdl)))
}

func runWebserver(log *slog.Logger) error {
	log.Info("Building routes...")

	mux := http.NewServeMux()

	for fv, resourceMap := range versionResourceMap {
		// get version resource list
		addHandler(log, mux, fmt.Sprintf("GET /%s/", fv.String()), handlerGetVersionResourceList(fv))

		for _, rscType := range resourceMap.ResourceTypes() {
			// get version resource bundle
			addHandler(log, mux, fmt.Sprintf("GET /%s/%s/", fv.String(), rscType), handlerGetResourceBundle(fv, rscType))

			// get specific version resource by id
			addHandler(log, mux, fmt.Sprintf("GET /%s/%s/{resource_id}/", fv.String(), rscType), handlerGetVersionResource(fv, rscType))
		}
	}

	// get version list
	addHandler(log, mux, "GET /{$}", handlerGetVersionList())

	log.Info("Webserver running", "addr", bindAddr)

	return http.ListenAndServe(bindAddr, mux)
}
