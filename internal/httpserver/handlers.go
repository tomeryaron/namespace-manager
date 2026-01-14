package httpserver

import (
	"encoding/json"
	"net/http"
)

// CreateNamespaceRequest represents the JSON request body for creating a namespace
type CreateNamespaceRequest struct {
	Name  string `json:"name"`  // Namespace name
	TTL   int    `json:"ttl"`   // Time to live in hours
	Owner string `json:"owner"` // Owner name
	Team  string `json:"team"`  // Team name
}

type DeleteNamespaceRequest struct {
	Name string `json:"name"`
}	

type NamespaceResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

func (s *Server) HandleCreateNamespaceRequest(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed. Use POST"))
		return
	}

	// Parse JSON request body
	var req CreateNamespaceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON: " + err.Error()))
		return
	}

	// Validate required fields
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required field: name"))
		return
	}
	if req.Owner == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required field: owner"))
		return
	}
	if req.Team == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required field: team"))
		return
	}
	if req.TTL <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("TTL must be greater than 0"))
		return
	}

	// Create the namespace
	err = s.kubeClient.CreateNamespace(req.Name, req.TTL, req.Owner, req.Team)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Namespace created successfully",
		"name":    req.Name,
	})
}

func (s *Server) HandleDeleteNamespaceRequest(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE method
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed. Use DELETE"))
		return
	}

	// Parse JSON request body
	var req DeleteNamespaceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON: " + err.Error()))
		return
	}

	// Delete the namespace
	err = s.kubeClient.DeleteNamespace(req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Namespace deleted successfully",
		"name":    req.Name,
	})
}

func (s *Server) HandleListNamespacesRequest(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed. Use GET"))
		return
	}

	// Get owner from query parameter (optional - empty string means list all)
	owner := r.URL.Query().Get("owner")

	// List the namespaces
	namespaces, err := s.kubeClient.ListNamespaces(owner)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(namespaces)
}