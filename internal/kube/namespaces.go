package kube

import (
	"time"
	"context" 
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceInfo represents namespace information with TTL details
type NamespaceInfo struct {
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	Team      string    `json:"team"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	TTL       int       `json:"ttl"` // TTL in hours
}

// CreateNamespace creates a namespace with TTL, owner, and team annotations
func (c *Client) CreateNamespace(name string, ttlHours int, owner string, team string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"owner": owner,
				"team": team,
				"expires_at": time.Now().Add(time.Duration(ttlHours) * time.Hour).Format(time.RFC3339),		// TTL in RFC3339 format for Kubernetes to parse
			},
		},
	}
	_, err := c.clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteNamespace(name string) error {
	// Delete the namespace (this returns immediately when Kubernetes accepts the request)
	err := c.clientset.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	
	// Create a context with a 30-second timeout - acts like a timer
	// ctx will signal when 30 seconds have passed
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // Clean up the timer when function exits
	
	// Poll loop: keep checking if namespace still exists
	for {
		// Try to get the namespace - if it doesn't exist, deletion succeeded
		_, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			// Namespace not found = successfully deleted
			return nil
		}
		
		// Check if our 30-second timer expired
		select {
		case <-ctx.Done():
			// Timer went off - timeout reached, return error
			return ctx.Err()
		default:
			// Timer still running - wait 500ms before checking again
			time.Sleep(500 * time.Millisecond)
		}
	}
}
 // ListNamespaces lists all namespaces, optionally filtered by owner
 func (c *Client) ListNamespaces(owner string) ([]NamespaceInfo, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	// Convert Kubernetes namespaces to NamespaceInfo
	var result []NamespaceInfo
	for _, ns := range namespaces.Items {
		// Extract annotations
		annotations := ns.ObjectMeta.Annotations
		nsOwner := annotations["owner"]
		team := annotations["team"]
		expiresAtStr := annotations["expires_at"]
		
		// Filter by owner if specified
		if owner != "" && nsOwner != owner {
			continue
		}
		
		// Parse expires_at to calculate TTL
		var expiresAt time.Time
		var ttl int
		if expiresAtStr != "" {
			expiresAt, err = time.Parse(time.RFC3339, expiresAtStr)
			if err == nil {
				// Calculate remaining TTL in hours
				remaining := time.Until(expiresAt)
				if remaining > 0 {
					ttl = int(remaining.Hours())
				}
			}
		}
		
		// Create NamespaceInfo
		info := NamespaceInfo{
			Name:      ns.ObjectMeta.Name,
			Owner:     nsOwner,
			Team:      team,
			CreatedAt: ns.ObjectMeta.CreationTimestamp.Time,
			ExpiresAt: expiresAt,
			TTL:       ttl,
		}
		
		result = append(result, info)
	}
	
	return result, nil
}

// // GetNamespaceTTL returns the remaining TTL for a namespace
// func (c *Client) GetNamespaceTTL(name string) (int, error) {
// 	// Implementation here
// }

// // GetNamespaceInfo returns full namespace information
// func (c *Client) GetNamespaceInfo(name string) (*NamespaceInfo, error) {
// 	// Implementation here
// }