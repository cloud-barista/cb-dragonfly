package metric

import (
	"reflect"

	"github.com/cloud-barista/cb-dragonfly/pkg/util"
)

// Cpu 메트릭
type Cpu struct {
	CpuGuest       float64 `json:"cpu_guest"`
	CpuGuestNice   float64 `json:"cpu_guest_nice"`
	CpuHardIrq     float64 `json:"cpu_hintr"`
	CpuIdle        float64 `json:"cpu_idle"`
	CpuIowait      float64 `json:"cpu_iowait"`
	CpuNice        float64 `json:"cpu_nice"`
	CpuSoftirq     float64 `json:"cpu_sintr"`
	CpuSteal       float64 `json:"cpu_steal"`
	CpuSystem      float64 `json:"cpu_system"`
	CpuUser        float64 `json:"cpu_user"`
	CpuUtilization float64 `json:"cpu_utilization"`
}

type Cpufreq struct {
	CpuSpeed float64 `json:"cpu_speed"`
}

func (c Cpufreq) GetField() []string {
	val := reflect.ValueOf(c)
	return util.GetFields(val)
}
func (c Cpu) GetField() []string {
	val := reflect.ValueOf(c)
	return util.GetFields(val)
}

// Memory 메트릭
type Memory struct {
	MemBuffers     float64 `json:"mem_buffers"`
	MemCached      float64 `json:"mem_cached"`
	MemFree        float64 `json:"mem_free"`
	MemShared      float64 `json:"mem_shared"`
	MemTotal       float64 `json:"mem_total"`
	MemUsed        float64 `json:"mem_used"`
	MemUtilization float64 `json:"mem_utilization"`
}

func (m Memory) GetField() []string {
	val := reflect.ValueOf(m)
	return util.GetFields(val)
}

// Disk 메트릭
type Disk struct {
	DiskUtilization string `json:"disk_utilization"`
	DiskTotal       string `json:"disk_total"`
	DiskUsed        string `json:"disk_used"`
	DiskFree        string `json:"disk_free"`
}

func (d Disk) GetField() []string {
	val := reflect.ValueOf(d)
	return util.GetFields(val)
}

// DiskIO 메트릭
type DiskIO struct {
	DiskReadBytes  string `json:"kb_read"`
	DiskWriteBytes string `json:"kb_written"`
	DIskReadIOPS   int64  `json:"ops_read"`
	DIskWriteIOPS  int64  `json:"ops_write"`
}

func (dio DiskIO) GetField() []string {
	val := reflect.ValueOf(dio)
	return util.GetFields(val)
}

// Network 메트릭
type Network struct {
	NetBytesIn   int64 `json:"bytes_in"`
	NetBytesOut  int64 `json:"bytes_out"`
	NetPacketIn  int64 `json:"pkts_in"`
	NetPacketOut int64 `json:"pkts_out"`
}

func (n Network) GetField() []string {
	val := reflect.ValueOf(n)
	return util.GetFields(val)
}

// MCK8SNode 메트릭
type MCK8SNode struct {
	CpuUsage                    int64 `json:"cpu_usage_nanocores"`
	MemUsage                    int64 `json:"memory_usage_bytes"`
	MemAvaiable                 int64 `json:"memory_available_bytes"`
	MemWorkingSet               int64 `json:"memory_working_set_bytes"`
	MemResidentSet              int64 `json:"memory_rss_bytes"`
	NetRXBytes                  int64 `json:"network_rx_bytes"`
	NetRXErrors                 int64 `json:"network_rx_errors"`
	NetTXBytes                  int64 `json:"network_tx_bytes"`
	NetTXErrors                 int64 `json:"network_tx_errors"`
	FSCapacityBytes             int64 `json:"fs_capacity_bytes"`
	FSUsedBytes                 int64 `json:"fs_used_bytes"`
	RuntimeImageFSCapacityBytes int64 `json:"runtime_image_fs_capacity_bytes"`
	RuntimeImageFSUsageBytes    int64 `json:"runtime_image_fs_usage_bytes"`
}

func (n MCK8SNode) GetField() []string {
	val := reflect.ValueOf(n)
	return util.GetFields(val)
}

// MCK8SPod 메트릭
type MCK8SPod struct {
	CpuUsage       int64 `json:"cpu_usage_nanocores"`
	MemUsage       int64 `json:"memory_usage_bytes"`
	MemWorkingSet  int64 `json:"memory_working_set_bytes"`
	MemResidentSet int64 `json:"memory_rss_bytes"`
	RootFSCapacity int64 `json:"rootfs_capacity_bytes"`
	RootFSUsed     int64 `json:"rootfs_used_bytes"`
	LogsFSCapacity int64 `json:"logsfs_capacity_bytes"`
	LogsFSUsage    int64 `json:"logsfs_usage_bytes"`
	NetRXBytes     int64 `json:"rx_bytes"`
	NetRXErrors    int64 `json:"rx_errors"`
	NetTXBytes     int64 `json:"tx_bytes"`
	NetTXErrors    int64 `json:"tx_errors"`
}

func (n MCK8SPod) GetField() []string {
	val := reflect.ValueOf(n)
	return util.GetFields(val)
}
