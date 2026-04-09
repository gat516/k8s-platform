package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
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

// handleResources reads Linux /proc files to compute live CPU, memory, disk I/O,
// and network throughput metrics. Two samples are taken 500 ms apart so that
// per-second rates can be calculated. If any /proc file is unreadable (e.g. running
// inside a restricted container) the corresponding metric is left at 0.
func (s *Server) handleResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	const sampleGap = 500 * time.Millisecond

	// First sample.
	cpu1, _ := readCPUStat()
	net1, _ := readNetStat()
	disk1, _ := readDiskStat()

	time.Sleep(sampleGap)

	// Second sample.
	cpu2, _ := readCPUStat()
	net2, _ := readNetStat()
	disk2, _ := readDiskStat()

	resp := resourcesResponse{}

	// CPU utilisation: (non-idle delta) / (total delta) * 100.
	if cpu1.total > 0 && cpu2.total > cpu1.total {
		idleDelta := float64(cpu2.idle - cpu1.idle)
		totalDelta := float64(cpu2.total - cpu1.total)
		resp.CPUPercent = roundTo1((1 - idleDelta/totalDelta) * 100)
	}

	// Memory: (total - available) / total * 100.
	mem, err := readMemInfo()
	if err == nil && mem.total > 0 {
		used := mem.total - mem.available
		resp.MemoryPercent = roundTo1(float64(used) / float64(mem.total) * 100)
	}

	// Disk I/O: approximate "busyness" as (io_ticks delta) / (elapsed_ms) * 100.
	// io_ticks counts milliseconds the device spent doing I/O in the measurement window.
	elapsedMS := sampleGap.Milliseconds()
	if disk1.ioTicks <= disk2.ioTicks && elapsedMS > 0 {
		resp.DiskIOPercent = roundTo1(float64(disk2.ioTicks-disk1.ioTicks) / float64(elapsedMS) * 100)
		if resp.DiskIOPercent > 100 {
			resp.DiskIOPercent = 100
		}
	}

	// Network: convert byte deltas over the sample window to Mbps.
	elapsedSec := sampleGap.Seconds()
	if net2.rxBytes >= net1.rxBytes {
		resp.NetInMbps = roundTo2(float64(net2.rxBytes-net1.rxBytes) * 8 / 1e6 / elapsedSec)
	}
	if net2.txBytes >= net1.txBytes {
		resp.NetOutMbps = roundTo2(float64(net2.txBytes-net1.txBytes) * 8 / 1e6 / elapsedSec)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ── /proc/stat helpers ────────────────────────────────────────────────────────

type cpuStat struct {
	idle  uint64
	total uint64
}

// readCPUStat parses the first "cpu" aggregate line from /proc/stat.
// The columns are: user nice system idle iowait irq softirq steal guest guest_nice.
func readCPUStat() (cpuStat, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStat{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return cpuStat{}, fmt.Errorf("unexpected /proc/stat format")
		}
		vals := make([]uint64, len(fields)-1)
		for i, f := range fields[1:] {
			vals[i], _ = strconv.ParseUint(f, 10, 64)
		}
		// Column index 3 is idle; index 4 is iowait (also counted as idle for our purposes).
		idle := vals[3]
		if len(vals) > 4 {
			idle += vals[4]
		}
		var total uint64
		for _, v := range vals {
			total += v
		}
		return cpuStat{idle: idle, total: total}, nil
	}
	return cpuStat{}, fmt.Errorf("cpu line not found in /proc/stat")
}

// ── /proc/meminfo helpers ─────────────────────────────────────────────────────

type memInfo struct {
	total     uint64 // kB
	available uint64 // kB
}

// readMemInfo parses MemTotal and MemAvailable from /proc/meminfo.
func readMemInfo() (memInfo, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return memInfo{}, err
	}
	defer f.Close()

	var m memInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			m.total = val
		case "MemAvailable:":
			m.available = val
		}
	}
	return m, nil
}

// ── /proc/diskstats helpers ───────────────────────────────────────────────────

type diskStat struct {
	ioTicks uint64 // milliseconds spent doing I/O (field 10, 0-indexed from field 3)
}

// readDiskStat aggregates io_ticks across all non-loop, non-ram block devices.
func readDiskStat() (diskStat, error) {
	f, err := os.Open("/proc/diskstats")
	if err != nil {
		return diskStat{}, err
	}
	defer f.Close()

	var total uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// /proc/diskstats has at least 14 fields per line.
		// Field index 2 is device name; field index 12 is io_ticks.
		if len(fields) < 13 {
			continue
		}
		name := fields[2]
		// Skip loop and ram devices; they aren't real block devices.
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}
		ticks, _ := strconv.ParseUint(fields[12], 10, 64)
		total += ticks
	}
	return diskStat{ioTicks: total}, nil
}

// ── /proc/net/dev helpers ─────────────────────────────────────────────────────

type netStat struct {
	rxBytes uint64
	txBytes uint64
}

// readNetStat aggregates RX/TX bytes across all non-loopback interfaces.
func readNetStat() (netStat, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return netStat{}, err
	}
	defer f.Close()

	var ns netStat
	scanner := bufio.NewScanner(f)
	// Skip the two header lines.
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		ns.rxBytes += rx
		ns.txBytes += tx
	}
	return ns, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func roundTo1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}

func roundTo2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
