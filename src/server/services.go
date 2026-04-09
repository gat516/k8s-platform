package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// serviceEntry is a single Deployment summary returned by GET /api/v1/services.
type serviceEntry struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	Image            string `json:"image"`
	ReplicasReady    int32  `json:"replicas_ready"`
	ReplicasDesired  int32  `json:"replicas_desired"`
	Status           string `json:"status"`
}

// handleServices lists all Deployments across every namespace and returns a
// condensed summary for each one.  When the k8s client is unavailable an empty
// array is returned so the frontend can show an empty state rather than error.
func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries := []serviceEntry{}

	if s.k8sClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		deployments, err := s.k8sClient.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, d := range deployments.Items {
				desired := int32(1)
				if d.Spec.Replicas != nil {
					desired = *d.Spec.Replicas
				}
				ready := d.Status.ReadyReplicas

				status := deploymentStatus(ready, desired)

				// Use the first container image as the representative image tag.
				image := ""
				if len(d.Spec.Template.Spec.Containers) > 0 {
					image = d.Spec.Template.Spec.Containers[0].Image
				}

				entries = append(entries, serviceEntry{
					Name:            d.Name,
					Namespace:       d.Namespace,
					Image:           image,
					ReplicasReady:   ready,
					ReplicasDesired: desired,
					Status:          status,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// deploymentStatus returns a human-readable status string consistent with the
// dashboard colour coding: "running", "pending", or "error".
func deploymentStatus(ready, desired int32) string {
	switch {
	case desired == 0:
		return "error"
	case ready >= desired:
		return "running"
	default:
		return "pending"
	}
}
