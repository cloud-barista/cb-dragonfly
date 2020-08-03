package config

import (
	"encoding/json"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/core"
	"github.com/mitchellh/mapstructure"
)

const (
	MonConfigKey = "/mon/config"
)

// 모니터링 정책 설정
func SetMonConfig(newMonConfig config.Monitoring) (*config.Monitoring, int, error) {
	config.GetInstance().SetMonConfig(newMonConfig)

	var monConfigMap map[string]interface{}
	err := mapstructure.Decode(config.GetInstance().Monitoring, &monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 모니터링 정책 etcd 저장
	err = core.CoreConfig.Etcd.WriteMetric(MonConfigKey, monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &config.GetInstance().Monitoring, http.StatusOK, nil
}

// 모니터링 정책 조회
func GetMonConfig() (*config.Monitoring, int, error) {
	// etcd에 저장된 모니터링 정책 조회
	etcdConfigVal, err := core.CoreConfig.Etcd.ReadMetric(MonConfigKey)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var monConfig config.Monitoring
	err = json.Unmarshal([]byte(etcdConfigVal.Value), &monConfig)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return &monConfig, http.StatusOK, nil
}

// 모니터링 정책 초기화
func ResetMonConfig() (*config.Monitoring, int, error) {
	defaultMonConfig := config.GetDefaultConfig().Monitoring

	var monConfigMap map[string]interface{}
	err := mapstructure.Decode(config.GetDefaultConfig().Monitoring, &monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 모니터링 정책 etcd 저장
	err = core.CoreConfig.Etcd.WriteMetric(MonConfigKey, monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &defaultMonConfig, http.StatusOK, nil
}
