package httpserver

import (
	"net/http"
	"namespace-manager/internal/kube"
)

// Server wraps the standard http.Server and adds routing capabilities
type Server struct {
	server    *http.Server  // The underlying HTTP server from Go's standard library
	mux       *http.ServeMux // Mux (multiplexer) - routes incoming requests to the right handler
	kubeClient *kube.Client  // Kubernetes client to perform namespace operations
}

// NewServer creates a new Server instance
// The mux is like a traffic director - it looks at the URL path and sends
// the request to the correct handler function
// kubeClient is the Kubernetes client that will be used by handlers to interact with Kubernetes
func NewServer(addr string, kubeClient *kube.Client) *Server {
	// Create a new mux (router) that will handle routing requests
	mux := http.NewServeMux()
	
	return &Server{
		server: &http.Server{
			Addr:    addr,        // Address to listen on (e.g., ":8080")
			Handler: mux,         // Tell the server to use our mux to route requests
		},
		mux:       mux,        // Store the mux so we can register routes on it later
		kubeClient: kubeClient, // Store the Kubernetes client so handlers can use it
	}
}

// RegisterRoute connects a URL path to a handler function
// When someone visits the path (e.g., "/"), the handler function will be called
// Example: RegisterRoute("/", handleRoot) means "when someone goes to /, call handleRoot"
func (s *Server) RegisterRoute(path string, handler http.HandlerFunc) {
	s.mux.HandleFunc(path, handler)
}

// ListenAndServe starts the HTTP server and begins listening for requests
// This is a blocking call - it will run until the server stops
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}