package types

// TODO: implements
type Metric struct{}

type Metrics struct {
	metrics []Metric
}

const (
	MONCONFIG           = "config"
	COLLECTORGROUPTOPIC = "collectorGroupTopic"
	TOPIC               = "topic"
)

const (
	NSID    = "nsId"
	MCISID  = "mcisId"
	VMID    = "vmId"
	OSTYPE  = "osType"
	CSPTYPE = "cspType"
)

const (
	CSP1 = "CSP1"
	CSP2 = "CSP2"
	CSP3 = "CSP3"
	CSP4 = "CSP4"
	CSP5 = "CSP5"
	CSP6 = "CSP6"
)
