package server

import (
	"fmt"
	"net/http"
)

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func (s *Server) handleArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Article")
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Archive")
}
