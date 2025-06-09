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

func handlerGetVersionResourceTypeList(fv FHIRVersion) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf("/%s", fv) && r.URL.Path != fmt.Sprintf("/%s/", fv) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		versionResourceMapMu.RLock()
		rm := versionResourceMap[fv]
		versionResourceMapMu.RUnlock()
		respondInKind(w, r, rm)
	}
}

func handlerGetVersionResourceBundle(fv FHIRVersion) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rscType := r.PathValue("rsc_type")
		rp := getRequestParams(r)

		versionResourceMapMu.RLock()
		rscs := versionResourceMap[fv].GetResourcesByType(rscType, rp.Count)
		versionResourceMapMu.RUnlock()

		if len(rscs) == 0 {
			http.Error(w, fmt.Sprintf("No resources of type %q for version %q found.", rscType, fv.String()), http.StatusNotFound)
			return
		}

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

func handlerGetVersionResource(fv FHIRVersion) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := getRequestLogger(r)
		rp := getRequestParams(r)
		if rp.Count != 0 {
			http.Error(w, "_count must be zero or undefined with specific resource ID", http.StatusBadRequest)
			return
		}

		rscType := r.PathValue("rsc_type")
		rscId := r.PathValue("rsc_id")

		versionResourceMapMu.RLock()
		rsc := versionResourceMap[fv].GetResource(rscType, rscId)
		versionResourceMapMu.RUnlock()

		if nil == rsc {
			log.Error("Resource not found", "rsc_id", rscId)
			http.Error(w, fmt.Sprintf("no version %q resource %q found with id %q", fv.String(), rscType, rscId), http.StatusNotFound)
			return
		}

		respondInKind(w, r, rsc)
	}
}

func addHandler(log *slog.Logger, mux *http.ServeMux, route string, hdl http.HandlerFunc) {
	log.Debug("Adding route handler", "route", route)

	mux.HandleFunc(route, middlewareEmbedLogger(log.With("route", route), middlewareParseRequestParams(hdl)))
}

func runWebserver(log *slog.Logger) error {
	log.Info("Building routes...")

	mux := http.NewServeMux()

	// get version list
	addHandler(log, mux, "GET /{$}", handlerGetVersionList())

	versionResourceMapMu.RLock()
	for fv := range versionResourceMap {
		// get version resource list
		addHandler(log, mux, fmt.Sprintf("GET /%s/", fv.String()), handlerGetVersionResourceTypeList(fv))
		addHandler(log, mux, fmt.Sprintf("GET /%s/{rsc_type}", fv.String()), handlerGetVersionResourceBundle(fv))
		addHandler(log, mux, fmt.Sprintf("GET /%s/{rsc_type}/{rsc_id}", fv.String()), handlerGetVersionResource(fv))
	}
	versionResourceMapMu.RUnlock()

	log.Info("Webserver running", "addr", bindAddr)

	return http.ListenAndServe(bindAddr, mux)
}
