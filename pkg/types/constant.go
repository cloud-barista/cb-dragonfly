package types

const (
	PushPolicy = "push"
	PullPolicy = "pull"
)

// CB-Store key
const (
	Agent                  = "/monitoring/agents/"
	MonConfig              = "/monitoring/configs"
	EventLog               = "/monitoring/eventLogs"
	CollectorPolicy        = "/monitoring/collectorPolicy"
	Topic                  = "/push/topic"
	CollectorTopicMap      = "/push/collectorTopicMap"
	MCK8STopic             = "/mck8s/push/topic"
	MCK8SCollectorTopicMap = "/mck8s/push/collectorTopicMap"
)

const (
	NsId    = "nsId"
	McisId  = "mcisId"
	VmId    = "vmId"
	OsType  = "osType"
	CspType = "cspType"
)

const (
	AgentCntCollectorPolicy = "AGENTCOUNT"
	CSPCollectorPolicy      = "CSP"
)

const (
	Alibaba     = "ALIBABA"
	Aws         = "AWS"
	Azure       = "AZURE"
	Cloudit     = "CLOUDIT"
	Cloudtwin   = "CLOUDTWIN"
	Docker      = "DOCKER"
	Gcp         = "GCP"
	Openstack   = "OPENSTACK"
	TotalCspCnt = 8
)

const (
	AgentPort            = 8888
	KafkaDefaultPort     = 9092
	InfluxDefaultPort    = 8086
	KapacitorDefaultPort = 9092
)

const (
	Dev     = "dev"
	Helm    = "helm"
	Compose = "compose"
)

const (
	TopicAdd = "TopicAdd"
	TopicDel = "TopicDel"
)

const (
	ConfigMapName       = "cb-dragonfly-collector-configmap"
	DeploymentName      = "cb-dragonfly-collector-"
	MCK8SConfigMapName  = "cb-dragonfly-mck8s-collector-configmap"
	MCK8SDeploymentName = "cb-dragonfly-mck8s-collector-"
)

const (
	LabelKey = "name"
	//Namespace = "dragonfly"

	MCISCollectorImage  = "cloudbaristaorg/cb-dragonfly:0.7.0-mcis-collector"
	MCK8SCollectorImage = "cloudbaristaorg/cb-dragonfly:0.7.0-mck8s-collector"
)

const (
	TBRestAPIURL = "http://localhost:1323/tumblebug"
)

const (
	CREATE = "create"
	DELETE = "delete"
)
