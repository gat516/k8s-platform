package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// resourcesResponse is the JSON body returned by GET /api/v1/resources.
type resourcesResponse struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	DiskIOPercent float64 `json:"disk_io_percent"`
	NetInMbps     float64 `json:"net_in_mbps"`
	NetOutMbps    float64 `json:"net_out_mbps"`
}

// promResult is the subset of the Prometheus instant-query API response we care about.
type promResult struct {
	Status string `json:"status"`
	Data   struct {
		Result []struct {
			Value [2]json.RawMessage `json:"value"` // [timestamp, "value"]
		} `json:"result"`
	} `json:"data"`
}

// prometheusClient wraps an HTTP client and a base URL for querying Prometheus.
var prometheusClient = &http.Client{Timeout: 5 * time.Second}

// queryPrometheus executes an instant PromQL query and returns the first scalar result.
// Returns 0 and a non-nil error if Prometheus is unreachable or returns no data.
func (s *Server) queryPrometheus(promQL string) (float64, error) {
	endpoint := s.cfg.PrometheusURL + "/api/v1/query?query=" + url.QueryEscape(promQL)

	resp, err := prometheusClient.Get(endpoint)
	if err != nil {
		return 0, fmt.Errorf("prometheus request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("reading prometheus response: %w", err)
	}

	var pr promResult
	if err := json.Unmarshal(body, &pr); err != nil {
		return 0, fmt.Errorf("parsing prometheus response: %w", err)
	}
	if pr.Status != "success" {
		return 0, fmt.Errorf("prometheus returned status %q", pr.Status)
	}
	if len(pr.Data.Result) == 0 {
		return 0, fmt.Errorf("prometheus returned no results for query: %s", promQL)
	}

	// value[1] is the string-encoded float.
	var raw string
	if err := json.Unmarshal(pr.Data.Result[0].Value[1], &raw); err != nil {
		return 0, fmt.Errorf("parsing prometheus value: %w", err)
	}
	return strconv.ParseFloat(raw, 64)
}

// handleResources queries Prometheus for node-level resource metrics and returns
// them as JSON. Falls back to 0 for any metric that cannot be retrieved, so a
// partial Prometheus outage does not fail the whole endpoint.
//
// All queries use node_exporter metrics shipped by kube-prometheus-stack.
// The 5m rate window smooths out short spikes without lagging too far behind.
func (s *Server) handleResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type query struct {
		promQL string
		target *float64
	}

	resp := resourcesResponse{}

	queries := []query{
		{
			// Node CPU: average utilisation across all cores, all modes except idle/iowait.
			`100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`,
			&resp.CPUPercent,
		},
		{
			// Node memory: fraction of total RAM that is not freely available.
			`(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`,
			&resp.MemoryPercent,
		},
		{
			// Block device I/O busyness: fraction of time at least one I/O was in flight.
			`avg(rate(node_disk_io_time_seconds_total{device!~"loop.*|ram.*"}[5m])) * 100`,
			&resp.DiskIOPercent,
		},
		{
			// Network receive throughput in Mbps (bits per second / 1e6).
			`sum(rate(node_network_receive_bytes_total{device!="lo"}[5m])) * 8 / 1e6`,
			&resp.NetInMbps,
		},
		{
			// Network transmit throughput in Mbps.
			`sum(rate(node_network_transmit_bytes_total{device!="lo"}[5m])) * 8 / 1e6`,
			&resp.NetOutMbps,
		},
	}

	for _, q := range queries {
		if val, err := s.queryPrometheus(q.promQL); err == nil {
			*q.target = roundTo2(val)
		}
		// Silently leave at 0 on error — partial data is better than a 500.
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func roundTo2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
