package sysinfo

import (
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type Snapshot struct {
	CPUPercent float64
	MemTotal   uint64
	MemUsed    uint64
	MemAvail   uint64
	SwapTotal  uint64
	SwapUsed   uint64
	DiskTotal  uint64
	DiskUsed   uint64
	DiskFree   uint64
	Procs      []ProcInfo
}

type ProcInfo struct {
	PID  int32
	Name string
	CPU  float64
	Mem  float32
	RSS  uint64
}

func TakeSnapshot(topN int) Snapshot {
	var s Snapshot

	if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
		s.CPUPercent = pcts[0]
	}

	if v, err := mem.VirtualMemory(); err == nil {
		s.MemTotal = v.Total
		s.MemUsed = v.Used
		s.MemAvail = v.Available
	}

	if sw, err := mem.SwapMemory(); err == nil {
		s.SwapTotal = sw.Total
		s.SwapUsed = sw.Used
	}

	if d, err := disk.Usage("/"); err == nil {
		s.DiskTotal = d.Total
		s.DiskUsed = d.Used
		s.DiskFree = d.Free
	}

	procs, err := process.Processes()
	if err != nil {
		return s
	}

	// Pass 1: seed the CPU baseline for every process.
	// p.Percent(0) with a zero interval records the current CPU ticks and
	// returns 0 on the very first call for each process object.  We must
	// call it once, wait, then call it again to get a real delta.
	for _, p := range procs {
		p.Percent(0) //nolint:errcheck // intentional seed; result is always 0
	}
	time.Sleep(500 * time.Millisecond)

	// Pass 2: collect actual measurements after the sample window.
	var infos []ProcInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		cpuPct, _ := p.Percent(0)
		memPct, _ := p.MemoryPercent()
		mi, _ := p.MemoryInfo()
		rss := uint64(0)
		if mi != nil {
			rss = mi.RSS
		}
		infos = append(infos, ProcInfo{
			PID: p.Pid, Name: name, CPU: cpuPct, Mem: memPct, RSS: rss,
		})
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].CPU > infos[j].CPU })
	if len(infos) > topN {
		infos = infos[:topN]
	}
	s.Procs = infos
	return s
}

func HumanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
