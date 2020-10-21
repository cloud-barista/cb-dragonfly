package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	InfluxDB       InfluxDB
	Etcd           Etcd
	CollectManager CollectManager
	APIServer      APIServer
	Monitoring     Monitoring
	Kapacitor      Kapacitor
}

type InfluxDB struct {
	EndpointUrl  string `json:"endpoint_url" mapstructure:"endpoint_url"`
	InternalPort int    `json:"internal_port" mapstructure:"internal_port"`
	ExternalPort int    `json:"external_port" mapstructure:"external_port"`
	Database     string
	UserName     string
	Password     string
}

type Etcd struct {
	EndpointUrl string `json:"endpoint_url" mapstructure:"endpoint_url"`
	ttl         int
}

type CollectManager struct {
	CollectorIP   string `json:"collector_ip" mapstructure:"collector_ip"`
	CollectorPort int    `json:"collector_port" mapstructure:"collector_port"`
	CollectorCnt  int    `json:"collector_count" mapstructure:"collector_count"`
}

type APIServer struct {
	Port int
}

type Monitoring struct {
	AgentInterval      int `json:"agent_interval" mapstructure:"agent_interval"`         // 모니터링 에이전트 수집주기
	AgentTTL           int `json:"agent_TTL" mapstructure:"agent_TTL"`                   // 모니터링 에이전트 데이터 TTL
	CollectorInterval  int `json:"collector_interval" mapstructure:"collector_interval"` // 모니터링 콜렉터 Aggregate 주기
	SchedulingInterval int `json:"schedule_interval" mapstructure:"schedule_interval"`   // 모니터링 콜렉터 스케줄링 주기 (스케일 인/아웃 로직 체크 주기)
	MaxHostCount       int `json:"max_host_count" mapstructure:"max_host_count"`         // 모니터링 콜렉터 수
}

type Kapacitor struct {
	EndpointUrl string `json:"endpoint_url" mapstructure:"endpoint_url"`
}

func (kapacitor Kapacitor) GetEndpointUrl() string {
	return kapacitor.EndpointUrl
}

var once sync.Once
var config Config

func GetInstance() *Config {
	once.Do(func() {
		loadConfigFromYAML(&config)
	})
	return &config
}

func GetDefaultConfig() *Config {
	var defaultMonConfig Config
	loadConfigFromYAML(&defaultMonConfig)
	return &defaultMonConfig
}

func (config *Config) SetMonConfig(newMonConfig Monitoring) {
	config.Monitoring = newMonConfig
}

func (config *Config) GetInfluxDBConfig() InfluxDB {
	return config.InfluxDB
}

func (config *Config) GetETCDConfig() Etcd {
	return config.Etcd
}

func (config *Config) GetKapacitorConfig() Kapacitor {
	return config.Kapacitor
}

func loadConfigFromYAML(config *Config) {
	configPath := os.Getenv("CBMON_ROOT") + "/conf"

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
