package types

const (
	PushPolicy = "push"
	PullPolicy = "pull"
)

// CB-Store key
const (
	Agent               = "/monitoring/agents"
	MonConfig           = "/monitoring/configs"
	Topic               = "/monitoring/topics"
	CollectorGroupTopic = "/monitoring/collectorGroupTopics"
	DeleteTopic         = "/monitoring/delTopics"
	EventLog            = "/monitoring/eventLogs"
)

const (
	NsId    = "nsId"
	McisId  = "mcisId"
	VmId    = "vmId"
	OsType  = "osType"
	CspType = "cspType"
)

const (
	AgentCnt = "AGENTCOUNT"
	CSP      = "CSP"
)

const (
	ALIBABA     = "ALIBABA"
	AWS         = "AWS"
	AZURE       = "AZURE"
	CLOUDIT     = "CLOUDIT"
	CLOUDTWIN   = "CLOUDTWIN"
	DOCKER      = "DOCKER"
	GCP         = "GCP"
	OPENSTACK   = "OPENSTACK"
	TotalCspCnt = 8
)

const (
	KafkaDefaultPort     = 9092
	InfluxDefaultPort    = 8086
	KapacitorDefaultPort = 9092
)
