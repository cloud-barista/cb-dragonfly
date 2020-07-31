package config

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/core"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
)

func SetMonConfig(agentInterval, collectorInterval, schedulingInterval, maxHostCnt, agentTTL int) (*config.MonConfig, int, error) {
	// 모니터링 정책 정보 설정
	config.GetInstance().SetMonConfig(agentInterval, collectorInterval, schedulingInterval, maxHostCnt, agentTTL)

	// TODO: 구조체 map[string]interface{} 타입으로 Unmarshal
	// TODO: 추후에 별도의 map 변환 함수 (toMap() 개발)
	fileBuffer := new(bytes.Buffer)
	err := json.NewEncoder(fileBuffer).Encode(config.GetInstance())
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Struct to Map 변환
	monConfigMap := map[string]interface{}{}
	monConfigBytes := fileBuffer.Bytes()
	err = json.Unmarshal(monConfigBytes, &monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 모니터링 정책 정보 etcd 저장
	err = core.CoreConfig.Etcd.WriteMetric("/mon/config", monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return config.GetInstance(), http.StatusOK, nil
}

func GetMonConfig() (*config.MonConfig, int, error) {
	// etcd 저장소에서 모니터링 정책 정보 조회
	etcdConfigVal, err := core.CoreConfig.Etcd.ReadMetric("/mon/config")
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 모니터링 정책 정보 구조체 매핑
	monConfig := config.MonConfig{}
	err = json.Unmarshal([]byte(etcdConfigVal.Value), &monConfig)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return &monConfig, http.StatusOK, nil
}

func ResetMonConfig() (*config.MonConfig, int, error) {
	defaultMonConfig := config.GetDefaultMonConfig()

	// TODO: 구조체 map[string]interface{} 타입으로 Unmarshal
	// TODO: 추후에 별도의 map 변환 함수 (toMap() 개발)
	fileBuffer := new(bytes.Buffer)
	err := json.NewEncoder(fileBuffer).Encode(defaultMonConfig)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Struct to Map 변환
	monConfigMap := map[string]interface{}{}
	monConfigBytes := fileBuffer.Bytes()
	err = json.Unmarshal(monConfigBytes, &monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 모니터링 정책 정보 etcd 저장
	err = core.CoreConfig.Etcd.WriteMetric("/mon/config", monConfigMap)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return defaultMonConfig, http.StatusOK, nil
}
