package config

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/localstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"

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

	mapstructure.Decode(config.GetInstance().Monitoring, &monConfigMap)
	for key, val := range monConfigMap {
		localstore.GetInstance().StorePut(types.MONCONFIG+"/"+key, fmt.Sprintf("%v", val))
	}
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &config.GetInstance().Monitoring, http.StatusOK, nil
}

// 모니터링 정책 조회
func GetMonConfig() (*config.Monitoring, int, error) {

	getValue := func(key string) int {
		value := localstore.GetInstance().StoreGet(types.MONCONFIG + "/" + key)
		result, _ := strconv.Atoi(value)
		return result
	}
	monConfig := config.Monitoring{
		MaxHostCount:      getValue("MaxHostCount"),
		AgentInterval:     getValue("AgentInterval"),
		CollectorInterval: getValue("CollectorInterval"),
	}

	if monConfig.CollectorInterval == -1 || monConfig.AgentInterval == -1 || monConfig.MaxHostCount == -1 {
		return nil, http.StatusInternalServerError, nil
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

	mapstructure.Decode(config.GetInstance().Monitoring, &monConfigMap)
	for key, val := range monConfigMap {
		localstore.GetInstance().StorePut(types.MONCONFIG+"/"+key, fmt.Sprintf("%v", val))
	}
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &defaultMonConfig, http.StatusOK, nil
}
