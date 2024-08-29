package services

import (
	"net/http"
	"net/http/pprof"
)

var DefaultProfilerURIPrefix = "/debug/pprof/"

func PprofMuxer(prefix string) *http.ServeMux {
	if prefix == "" {
		prefix = DefaultProfilerURIPrefix
	}

	muxer := http.NewServeMux()
	muxer.HandleFunc(prefix, pprof.Index)
	muxer.HandleFunc(prefix+"cmdline", pprof.Cmdline)
	muxer.HandleFunc(prefix+"profile", pprof.Profile)
	muxer.HandleFunc(prefix+"symbol", pprof.Symbol)
	muxer.HandleFunc(prefix+"trace", pprof.Trace)

	return muxer
}
