package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// clusterResponse is the JSON body returned by GET /api/v1/cluster.
type clusterResponse struct {
	Nodes         int    `json:"nodes"`
	PodsRunning   int    `json:"pods_running"`
	PodsTotal     int    `json:"pods_total"`
	Restarts24h   int    `json:"restarts_24h"`
	K3sVersion    string `json:"k3s_version"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

// handleCluster returns a summary of cluster health: node count, pod counts,
// restart count in the last 24 h, k3s version, and the age of the oldest node.
// When the k8s client is unavailable it returns zeroed mock data so the frontend
// can still render without crashing.
func (s *Server) handleCluster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := clusterResponse{}

	if s.k8sClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		nodes, err := s.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err == nil {
			resp.Nodes = len(nodes.Items)

			// Use the creation timestamp of the oldest node as proxy for cluster uptime.
			for _, n := range nodes.Items {
				age := time.Since(n.CreationTimestamp.Time).Seconds()
				if int64(age) > resp.UptimeSeconds {
					resp.UptimeSeconds = int64(age)
				}
				// Harvest the kubelet / k3s version from the first ready node.
				if resp.K3sVersion == "" {
					resp.K3sVersion = n.Status.NodeInfo.KubeletVersion
				}
			}
		}

		cutoff := time.Now().Add(-24 * time.Hour)

		pods, err := s.k8sClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err == nil {
			resp.PodsTotal = len(pods.Items)
			for _, p := range pods.Items {
				// Count running pods.
				if p.Status.Phase == "Running" {
					resp.PodsRunning++
				}
				// Sum container restarts that occurred within the last 24 h.
				for _, cs := range p.Status.ContainerStatuses {
					if cs.LastTerminationState.Terminated != nil {
						if cs.LastTerminationState.Terminated.FinishedAt.After(cutoff) {
							resp.Restarts24h++
						}
					}
				}
			}
		}
	}

	if resp.K3sVersion == "" {
		resp.K3sVersion = "unknown"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
