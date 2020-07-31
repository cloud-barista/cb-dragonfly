package config

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InfluxDB struct {
		EndpointUrl string `yaml:"endpoint_url"`
		Database    string `yaml:"database"`
		UserName    string `yaml:"user_name"`
		Password    string `yaml:"password"`
	} `yaml:"influxdb"`
	Etcd struct {
		EndpointUrl string `yaml:"endpoint_url"`
	} `yaml:"etcd"`
	CollectManager struct {
		CollectorIP   string `yaml:"collector_ip"`
		CollectorPort int    `yaml:"collector_port"`
		CollectorCnt  int    `yaml:"collector_count"`
	} `yaml:"collect_manager"`
	APIServer struct {
		Port int `yaml:"port"`
	} `yaml:"api_server"`
	Monitoring struct {
		AgentTtl          int `yaml:"agent_TTL"`
		AgentInterval     int `yaml:"agent_interval"`
		CollectorInterval int `yaml:"collector_interval"`
		ScheduleInterval  int `yaml:"schedule_interval"`
		MaxHostCount      int `yaml:"max_host_count"`
	} `yaml:"monitoring"`
}

type MonConfig struct {
	Monitoring struct {
		AgentInterval      int `json:"agent_interval"`     // 모니터링 에이전트 수집주기
		CollectorInterval  int `json:"collector_interval"` // 모니터링 콜렉터 Aggregate 주기
		SchedulingInterval int `json:"schedule_interval"`  // 모니터링 콜렉터 스케줄링 주기 (스케일 인/아웃 로직 체크 주기)
		MaxHostCount       int `json:"max_host_count"`     // 모니터링 콜렉터 수
		AgentTTL           int `json:"agent_TTL"`          // 모니터링 에이전트 데이터 TTL
	} `yaml:"monitoring"`
}

func (m *MonConfig) SetMonConfig(agentInterval, collectorInterval, schedulingInterval, maxHostCount, agentTTL int) {
	m.Monitoring.AgentInterval = agentInterval
	m.Monitoring.CollectorInterval = collectorInterval
	m.Monitoring.SchedulingInterval = schedulingInterval
	m.Monitoring.MaxHostCount = maxHostCount
	m.Monitoring.AgentTTL = agentTTL
}

func (m *MonConfig) GetAgentTTL() int {
	return m.Monitoring.AgentTTL
}

func (m *MonConfig) GetAgentInterval() int {
	return m.Monitoring.AgentInterval
}

func (m *MonConfig) GetCollectorInterval() int {
	return m.Monitoring.CollectorInterval
}

func (m *MonConfig) GetSchedulingInterval() int {
	return m.Monitoring.SchedulingInterval
}

func (m *MonConfig) GetMaxHostCount() int {
	return m.Monitoring.MaxHostCount
}

var monConfig *MonConfig
var once sync.Once

func GetDefaultMonConfig() *MonConfig {
	defaultMonConfig := MonConfig{}
	loadConfigFromYAML(&defaultMonConfig)
	return &defaultMonConfig
}

func GetInstance() *MonConfig {
	once.Do(func() {
		monConfig = &MonConfig{}
		loadConfigFromYAML(monConfig)
	})
	return monConfig
}

func loadConfigFromYAML(monConfig *MonConfig) {
	configPath := os.Getenv("CBMON_ROOT") + "/conf/config.yaml"

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Error("Failed to read configuration file in: ", configPath)
		return
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		logrus.Error("Failed to unmarshal configuration file")
		return
	}

	//monConfig = config.Monitoring

	return
}
