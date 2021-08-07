package httpapi

import (
	"context"
	"net/http"
)

type Server struct {
	address string
	mux     *http.ServeMux
	server  *http.Server
}

func NewServer(address string) *Server {
	mux := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	server := &Server{
		address: address,
		mux:     mux,
		server:  httpServer,
	}

	return server
}

func (s *Server) Register(path string, handler http.HandlerFunc) *Server {
	s.mux.Handle(path, handler)

	return s
}

func (s *Server) Listen() error {
	err := s.server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			return nil
		}

		return err
	}

	return nil
}

func (s *Server) Close() error {
	return s.server.Shutdown(context.Background())
}
