package types

type MetricType string

const (
	CPU     MetricType = "cpu"
	CPUFREQ MetricType = "cpufreq"
	DISK    MetricType = "disk"
	DISKIO  MetricType = "diskio"
	NETWORK MetricType = "network"
)

func (m MetricType) ToString() string {
	switch m {
	case CPU:
		return "cpu"
	case CPUFREQ:
		return "cpufreq"
	case DISK:
		return "disk"
	case DISKIO:
		return "diskio"
	case NETWORK:
		return "network"
	default:
		return ""
	}
}
