package diagnostics

import (
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// These pin the gopsutil v4 surface this app actually depends on. A major
// bump that compiles but stops returning usable data would otherwise only
// show up as an empty dashboard at runtime.
//
// Assertions are deliberately loose: they check the call succeeds and the
// shape is sane, not specific values, so they stay stable across machines
// and CI runners.

func TestGopsutilDiskUsage(t *testing.T) {
	usage, err := disk.Usage("/")
	if err != nil {
		t.Fatalf("disk.Usage(%q) returned an error: %v", "/", err)
	}
	if usage.Total == 0 {
		t.Error("disk.Usage reported a total of 0 bytes for the root filesystem")
	}
	if usage.UsedPercent < 0 || usage.UsedPercent > 100 {
		t.Errorf("disk.Usage reported UsedPercent %.2f, want 0..100", usage.UsedPercent)
	}
}

func TestGopsutilVirtualMemory(t *testing.T) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		t.Fatalf("mem.VirtualMemory returned an error: %v", err)
	}
	if vm.Total == 0 {
		t.Error("mem.VirtualMemory reported a total of 0 bytes")
	}
	if vm.UsedPercent < 0 || vm.UsedPercent > 100 {
		t.Errorf("mem.VirtualMemory reported UsedPercent %.2f, want 0..100", vm.UsedPercent)
	}
}

func TestGopsutilHostInfo(t *testing.T) {
	info, err := host.Info()
	if err != nil {
		t.Fatalf("host.Info returned an error: %v", err)
	}
	if info.OS == "" {
		t.Error("host.Info returned an empty OS")
	}

	if _, err := host.Uptime(); err != nil {
		t.Errorf("host.Uptime returned an error: %v", err)
	}
}

func TestGopsutilCPUPercent(t *testing.T) {
	// A zero interval reports usage since boot and does not block.
	pct, err := cpu.Percent(0, false)
	if err != nil {
		t.Fatalf("cpu.Percent returned an error: %v", err)
	}
	if len(pct) == 0 {
		t.Fatal("cpu.Percent returned no samples")
	}
	if pct[0] < 0 || pct[0] > 100 {
		t.Errorf("cpu.Percent returned %.2f, want 0..100", pct[0])
	}
}

func TestGopsutilNetCounters(t *testing.T) {
	counters, err := net.IOCounters(false)
	if err != nil {
		t.Fatalf("net.IOCounters returned an error: %v", err)
	}
	if len(counters) == 0 {
		t.Error("net.IOCounters returned no aggregate entry")
	}

	if _, err := net.Interfaces(); err != nil {
		t.Errorf("net.Interfaces returned an error: %v", err)
	}
}

func TestCheckDiskReportsRootFilesystem(t *testing.T) {
	r := CheckDisk()

	if r.Category != "Disk" {
		t.Errorf("CheckDisk category = %q, want %q", r.Category, "Disk")
	}
	if r.Status == Critical && r.Summary == "Unable to read disk usage" {
		t.Fatalf("CheckDisk could not read the root filesystem: %+v", r)
	}
	if r.Summary == "" {
		t.Error("CheckDisk returned an empty summary")
	}
	if len(r.Details) != 3 {
		t.Errorf("CheckDisk returned %d detail lines, want 3 (total, used, free)", len(r.Details))
	}
}
