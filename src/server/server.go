package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"presence/app"
	"presence/logger"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	app       *app.App
	srv       *http.Server
	srvtls    *http.Server
	accessLog *logger.Logger
	done      chan struct{}
	once      sync.Once
}

func New(a *app.App) *Server {
	return &Server{
		app:       a,
		accessLog: logger.NewLogger(),
	}
}

func (s *Server) newHTTPServer(port uint) *http.Server {
	r := mux.NewRouter()

	r.HandleFunc("/", s.handleHome)
	r.HandleFunc("/archive", s.handleArchive)
	r.HandleFunc("/{page:[0-9]+}/", s.handleHome)
	r.HandleFunc("/{slug:[a-zA-Z0-9_-]+}", s.handleArticle)

	h := s.withLogging(handlers.CompressHandler(s.withCommonHeaders(r)))

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		Handler:      h,
	}
}

func (s *Server) newTLSRedirectServer() *http.Server {
	redirect := func(w http.ResponseWriter, r *http.Request) {
		u := url.URL{
			Scheme: "https",
			Host:   r.Host,
			Path:   r.URL.Path,
		}
		http.Redirect(w, r, u.String(), 301)
	}
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", s.app.Config.Port),
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		Handler:      http.HandlerFunc(redirect),
	}
}

// initServers initializes the servers based on app config.
func (s *Server) initServers() error {
	if s.app.Config.Port == 0 {
		return fmt.Errorf("HTTP port is not set; check your configuration")
	}
	if s.app.Config.PortTLS != 0 {
		if s.app.Config.TLSKey == "" || s.app.Config.TLSCert == "" {
			return fmt.Errorf("TLS key and certificate must be set to handle HTTPS requests")
		}
		s.srvtls = s.newHTTPServer(s.app.Config.PortTLS)
		if s.app.Config.ForceTLS {
			s.srv = s.newTLSRedirectServer()
		}
	}
	if s.srv == nil {
		s.srv = s.newHTTPServer(s.app.Config.Port)
	}
	return nil
}

func (s *Server) getRemoteAddressForRequest(r *http.Request) string {
	proxies := int(s.app.Config.ProxyCount)
	if proxies > 0 {
		h := r.Header.Get("X-Forwarded-For")
		if h != "" {
			clients := strings.Split(h, ",")
			if proxies > len(clients) {
				return clients[0]
			}
			return clients[len(clients)-proxies]
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		panic(err)
	}
	return host
}

func (s *Server) Run() error {
	if s.done != nil {
		panic("Server.Run called twice")
	}
	if err := s.initServers(); err != nil {
		return err
	}

	s.done = make(chan struct{})
	errch := make(chan error)

	go func() {
		log.Printf("starting HTTP server at :%d...", s.app.Config.Port)
		err := s.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v\n", err)
			errch <- err
		}
	}()

	if s.srvtls != nil {
		go func() {
			log.Printf("starting HTTPS server at :%d...", s.app.Config.PortTLS)
			err := s.srv.ListenAndServeTLS(s.app.Config.TLSCert, s.app.Config.TLSKey)
			if err != nil && err != http.ErrServerClosed {
				log.Printf("server error: %v\n", err)
				errch <- err
			}
		}()
	}

	// Assume things are running smoothly one second in with no errors.
	check := time.AfterFunc(1*time.Second, func() {
		log.Printf("services are ready")
	})

	// Block until App is closed or there is an error during startup.
	select {
	case err := <-errch:
		check.Stop()
		s.Close()
		return err
	case <-s.done:
		return nil
	}
}

func (s *Server) close() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var wg sync.WaitGroup

	log.Println("waiting for connections to finish...")

	go func() {
		wg.Add(1)
		if err := s.srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown: %v", err)
		}
		wg.Done()
	}()

	if s.srvtls != nil {
		go func() {
			wg.Add(1)
			if err := s.srvtls.Shutdown(ctx); err != nil {
				log.Printf("HTTPS server shutdown: %v", err)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	s.accessLog.Close()
	close(s.done)
}

func (s *Server) Close() {
	if s.done == nil {
		return
	}
	s.once.Do(func() {
		s.close()
	})
}
