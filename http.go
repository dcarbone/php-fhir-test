package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type RequestParams struct {
	Format string
	Count  int
}

func logMiddlewareHandler(log *slog.Logger, hdl http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing request...", "method", r.Method, "url", r.URL)
		hdl(w, r)
	}
}

func versionListHandler(log *slog.Logger) http.HandlerFunc {
	return logMiddlewareHandler(log, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" && r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		if err := enc.Encode(versionList()); err != nil {
			log.Error("Error encoding version list", "err", err)
			http.Error(w, "Error encoding version list", http.StatusInternalServerError)
		}
	})
}

func versionResourceListHandler(log *slog.Logger, fv string) http.HandlerFunc {
	return logMiddlewareHandler(log, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != fmt.Sprintf("/%s", fv) && r.URL.Path != fmt.Sprintf("/%s/", fv) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		if err := enc.Encode(versionResourceList(fv)); err != nil {
			log.Error("Error encoding version resource list", "version", fv, "err", err)
			http.Error(w, fmt.Sprintf("Error encoding version %q resource list", fv), http.StatusInternalServerError)
		}
	})
}

func resourceTypeListHandler(log *slog.Logger, fv, resType string) http.HandlerFunc {
	type bundle struct {
		ResourceType string      `json:"resourceType"`
		Entry        []*Resource `json:"entry"`
	}

	return logMiddlewareHandler(log, func(w http.ResponseWriter, r *http.Request) {
		rp, err := parseRequestParams(r)
		if err != nil {
			log.Error("Error parsing query params", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if rp.Count < 0 {
			http.Error(w, fmt.Sprintf("_count must be non-negative, saw %d", rp.Count), http.StatusBadRequest)
			return
		}

		cnt := rp.Count
		if cnt == 0 {
			cnt = len(resourceMap[fv][resType])
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		out := bundle{
			ResourceType: "Bundle",
			Entry:        make([]*Resource, cnt),
		}

		for i := 0; i < cnt; i++ {
			out.Entry[i] = resourceMap[fv][resType][i]
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(out); err != nil {
			log.Error("Error encoding version resource bundle", "version", fv, "resourceType", resType, "err", err)
			http.Error(w, fmt.Sprintf("error encoding version %q resource %q bundle", fv, resType), http.StatusInternalServerError)
		}
	})
}

func resourceHandler(log *slog.Logger, fv, resType string, i int) http.HandlerFunc {
	return logMiddlewareHandler(log, func(w http.ResponseWriter, r *http.Request) {
		rp, err := parseRequestParams(r)
		if err != nil {
			log.Error("Error parsing query params", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if rp.Count != 0 {
			http.Error(w, "_count must be zero or undefined with specific resource ID", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		res := resourceMap[fv][resType][i]
		if err := enc.Encode(resourceMap[fv][resType][i]); err != nil {
			log.Error("Error encoding version resource", "version", fv, "resourceType", resType, "resourceID", res.ID, "err", err)
			http.Error(w, fmt.Sprintf("error encoding version %q resource %q id %q", fv, resType, res.ID), http.StatusInternalServerError)
		}
	})
}

func runWebserver(log *slog.Logger) error {
	log.Info("Building routes...")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", versionListHandler(log))

	for fv, resourceTypes := range resourceMap {
		mux.HandleFunc(fmt.Sprintf("GET /%s", fv), versionResourceListHandler(log, fv))
		mux.HandleFunc(fmt.Sprintf("GET /%s/", fv), versionResourceListHandler(log, fv))

		for resType, resources := range resourceTypes {
			mux.HandleFunc(fmt.Sprintf("GET /%s/%s", fv, resType), resourceTypeListHandler(log, fv, resType))
			mux.HandleFunc(fmt.Sprintf("GET /%s/%s/", fv, resType), resourceTypeListHandler(log, fv, resType))

			for i, res := range resources {
				if res.ID == "" {
					continue
				}
				mux.HandleFunc(fmt.Sprintf("GET /%s/%s/%s", fv, resType, res.ID), resourceHandler(log, fv, resType, i))
				mux.HandleFunc(fmt.Sprintf("GET /%s/%s/%s/", fv, resType, res.ID), resourceHandler(log, fv, resType, i))
			}
		}
	}

	log.Info("Webserver running", "addr", bindAddr)

	return http.ListenAndServe(bindAddr, mux)
}
