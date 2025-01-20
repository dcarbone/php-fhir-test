package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func resourceTypeListHandler(log *slog.Logger, fv, resType string) http.HandlerFunc {
	type bundle struct {
		ResourceType string      `json:"resourceType"`
		Entry        []*Resource `json:"entry"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		out := bundle{
			ResourceType: "Bundle",
			Entry:        make([]*Resource, len(resourceMap[fv][resType])),
		}
		for i, res := range resourceMap[fv][resType] {
			out.Entry[i] = res
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(out); err != nil {
			log.Error("Error encoding version resource bundle", "version", fv, "resourceType", resType, "err", err)
			http.Error(w, fmt.Sprintf("error encoding version %q resource %q bundle", fv, resType), http.StatusInternalServerError)
		}
	}
}

func resourceHandler(log *slog.Logger, fv, resType string, i int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		res := resourceMap[fv][resType][i]
		if err := enc.Encode(resourceMap[fv][resType][i]); err != nil {
			log.Error("Error encoding version resource", "version", fv, "resourceType", resType, "resourceID", res.ID, "err", err)
			http.Error(w, fmt.Sprintf("error encoding version %q resource %q id %q", fv, resType, res.ID), http.StatusInternalServerError)
		}
	}
}

func runWebserver(log *slog.Logger) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		if err := enc.Encode(versionList()); err != nil {
			log.Error("Error encoding version list", "err", err)
			http.Error(w, "Error encoding version list", http.StatusInternalServerError)
		}
	})

	for fv, resourceTypes := range resourceMap {
		mux.HandleFunc(fmt.Sprintf("GET /%s", fv), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			enc := json.NewEncoder(w)
			if err := enc.Encode(versionResourceList(fv)); err != nil {
				log.Error("Error encoding version resource list", "version", fv, "err", err)
				http.Error(w, fmt.Sprintf("Error encoding version %q resource list", fv), http.StatusInternalServerError)
			}
		})

		for resType, resources := range resourceTypes {
			mux.HandleFunc(fmt.Sprintf("GET /%s/%s", fv, resType), resourceTypeListHandler(log, fv, resType))

			for i, res := range resources {
				if res.ID == "" {
					continue
				}
				mux.HandleFunc(fmt.Sprintf("GET /%s/%s/%s", fv, resType, res.ID), resourceHandler(log, fv, resType, i))
			}
		}
	}

	log.Info("Webserver running", "addr", bindAddr)

	return http.ListenAndServe(bindAddr, mux)
}
