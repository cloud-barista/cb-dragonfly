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
}

type InfluxDB struct {
	EndpointUrl string
	Database    string
	UserName    string
	Password    string
}

type Etcd struct {
	EndpointUrl string
}

type CollectManager struct {
	CollectorIP   string
	CollectorPort int
	CollectorCnt  int
}

type APIServer struct {
	Port int
}

type Monitoring struct {
	AgentInterval      int `json:"agent_interval"`     // 모니터링 에이전트 수집주기
	CollectorInterval  int `json:"collector_interval"` // 모니터링 콜렉터 Aggregate 주기
	SchedulingInterval int `json:"schedule_interval"`  // 모니터링 콜렉터 스케줄링 주기 (스케일 인/아웃 로직 체크 주기)
	MaxHostCount       int `json:"max_host_count"`     // 모니터링 콜렉터 수
	AgentTTL           int `json:"agent_TTL"`          // 모니터링 에이전트 데이터 TTL
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
