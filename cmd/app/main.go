package main

import (
	"log"
	"namespace-manager/internal/httpserver"
	"namespace-manager/internal/kube"
)

func main() {
	// Create Kubernetes client first - this will be used by handlers to interact with Kubernetes
	kubeClient, err := kube.NewClient()
	if err != nil {
		log.Fatalf("Failed to create kube client: %v", err)
	}
	
	// Create a new server that will listen on port 8080
	// Pass the Kubernetes client so handlers can use it
	server := httpserver.NewServer(":8080", kubeClient)
	
	// Register the root path "/" with the HandleRoot handler
	// This means: when someone visits http://localhost:8080/, 
	// the HandleRoot function will be called to handle the request
	server.RegisterRoute("/api/namespaces/create", server.HandleCreateNamespaceRequest) 

	server.RegisterRoute("/api/namespaces/delete", server.HandleDeleteNamespaceRequest)

	server.RegisterRoute("/api/namespaces/list", server.HandleListNamespacesRequest)
	
	// Start the server and begin listening for incoming HTTP requests
	// This will block (keep running) until the server stops
	log.Println("Server starting on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}