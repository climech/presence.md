package server

import (
	"net/http"
	"time"
)

func (s *Server) withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10*1024)

		if s.app.Config.ForceTLS {
			w.Header().Add(
				"Strict-Transport-Security",
				"max-age=63072000; includeSubDomains",
			)
		}

		next.ServeHTTP(w, r)
	})
}

type statusCodeRecorder struct {
	http.ResponseWriter
	http.Hijacker
	StatusCode int
}

func (r *statusCodeRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()

		// Hijack to record response status and duration.
		hijacker, _ := w.(http.Hijacker)
		w = &statusCodeRecorder{
			ResponseWriter: w,
			Hijacker:       hijacker,
		}

		// Log access after request is processed.
		defer func() {
			statusCode := w.(*statusCodeRecorder).StatusCode
			if statusCode == 0 {
				statusCode = 200
			}
			remoteAddr := s.getRemoteAddressForRequest(r)
			duration := time.Since(timeStart)
			s.accessLog.Printf(
				"%v %v %v%v %v %v (%v)\n",
				remoteAddr,
				r.Method,
				r.Host,
				r.RequestURI,
				r.Proto,
				statusCode,
				duration,
			)
		}()

		next.ServeHTTP(w, r)
	})
}
