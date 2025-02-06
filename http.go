package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

var (
	errInvalidFormat = errors.New("invalid format")
)

type RequestParams struct {
	Format string
	Count  int
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
		return rp, fmt.Errorf("%w: %s", errInvalidFormat, format)
	}

	countstr := r.URL.Query().Get("_count")
	if countstr != "" {
		if rp.Count, err = strconv.Atoi(countstr); err != nil {
			return rp, fmt.Errorf("invalid value provided to _count: %s", countstr)
		}
	}

	return rp, nil
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

func logRequestMiddleware(log *slog.Logger, hdl http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Processing request...", "method", r.Method, "url", r.URL)
		hdl(w, r)
	}
}

func versionListHandler(log *slog.Logger) http.HandlerFunc {
	return logRequestMiddleware(log, func(w http.ResponseWriter, r *http.Request) {
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

func versionResourceListHandler(log *slog.Logger, fv string) http.HandlerFunc {
	return logRequestMiddleware(log, func(w http.ResponseWriter, r *http.Request) {
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

func resourceBundleHandler(log *slog.Logger, fv, resType string) http.HandlerFunc {
	return logRequestMiddleware(log, func(w http.ResponseWriter, r *http.Request) {
		rp, err := parseRequestParams(r)
		if err != nil {
			if errors.Is(err, errInvalidFormat) {
				log.Error("Invalid format requested", "format", rp.Format, "err", err)
				http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			} else {
				log.Error("Error parsing query params", "err", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		} else if rp.Count < 0 {
			http.Error(w, fmt.Sprintf("_count must be non-negative, saw %d", rp.Count), http.StatusBadRequest)
			return
		}

		cnt := rp.Count
		if cnt == 0 {
			cnt = len(resourceMap[fv][resType])
		}

		out := Bundle{
			ResourceType: "Bundle",
			Entry:        make([]BundleEntry, cnt),
		}
		for i := 0; i < cnt; i++ {
			out.Entry[i] = BundleEntry{Resource: resourceMap[fv][resType][i]}
		}

		respondInKind(log, rp, w, fv, out)
	})
}

func resourceHandler(log *slog.Logger, fv, resType string, i int) http.HandlerFunc {
	return logRequestMiddleware(log, func(w http.ResponseWriter, r *http.Request) {
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

		respondInKind(log, rp, w, fv, resourceMap[fv][resType][i])
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
			mux.HandleFunc(fmt.Sprintf("GET /%s/%s", fv, resType), resourceBundleHandler(log, fv, resType))
			mux.HandleFunc(fmt.Sprintf("GET /%s/%s/", fv, resType), resourceBundleHandler(log, fv, resType))

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
