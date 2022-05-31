package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/storage/cbstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"strings"
)

// AgentType 에이전트 동작 메커니즘 유형 (Push, Pull)
type AgentType string

const (
	Push AgentType = "push"
	Pull AgentType = "pull"
)

// AgentState 에이전트 설치 상태 (설치, 제거)
type AgentState string

const (
	Enable  AgentState = "enable"
	Disable AgentState = "disable"
)

// AgentHealth 에이전트 구동 상태 (정상, 비정상)
type AgentHealth string

const (
	Healthy   AgentHealth = "healthy"
	Unhealthy AgentHealth = "unhealthy"
)

// AgentInfo 에이전트 상세 정보
type AgentInfo struct {
	NsId                  string `json:"ns_id"`
	McisId                string `json:"mcis_id"`
	VmId                  string `json:"vm_id"`
	CspType               string `json:"csp_type"`
	AgentType             string `json:"agent_type"`
	AgentState            string `json:"agent_state"`
	AgentHealth           string `json:"agent_health"`
	AgentUnhealthyRespCnt int    `json:"agent_unhealthy_resp_cnt"`
	PublicIp              string `json:"public_ip"`
	ServiceType           string `json:"service_type"`
	McksID                string `json:"mcks_id"`
	APIServerURL          string `json:"apiserver_url"`
	ServerCA              string `json:"server_ca"`
	ClientCA              string `json:"client_ca"`
	ClientKey             string `json:"client_key"`
	ClientToken           string `json:"client_token"`
}

func MakeAgentUUID(info AgentInstallInfo) string {
	mcksType := strings.EqualFold(info.ServiceType, MCKS) || strings.EqualFold(info.ServiceType, MCKSAGENT_TYPE) || strings.EqualFold(info.ServiceType, MCKSAGENT_SHORTHAND_TYPE)
	if mcksType {
		return fmt.Sprintf("%s_%s_%s", info.NsId, info.ServiceType, info.McksID)
	}
	return fmt.Sprintf("%s_%s_%s_%s_%s", info.NsId, info.ServiceType, info.McisId, info.VmId, info.CspType)
}

// AgentListManager 에이전트 목록 관리
func DeleteAgent(info AgentInstallInfo) error {
	agentUUID := MakeAgentUUID(info)
	if err := cbstore.GetInstance().StoreDelete(types.Agent + agentUUID); err != nil {
		return err
	}
	return nil
}

func ListAgent() (map[string]AgentInfo, error) {
	agentList := map[string]AgentInfo{}
	agentListByteMap := cbstore.GetInstance().StoreGetListMap(types.Agent, true)

	if len(agentListByteMap) != 0 {
		for uuid, bytes := range agentListByteMap {
			agent := AgentInfo{}
			if err := json.Unmarshal([]byte(bytes), &agent); err != nil {
				return nil, errors.New(fmt.Sprintf("failed to convert agent list, error=%s", err))
			}
			agentList[uuid] = agent
		}
	}
	return agentList, nil
}

func GetAgent(info AgentInstallInfo) (*AgentInfo, error) {
	agentUUID := MakeAgentUUID(info)
	agentInfo := AgentInfo{}
	agentInfoStr := cbstore.GetInstance().StoreGet(fmt.Sprintf(types.Agent + agentUUID))

	if agentInfoStr == "" {
		return nil, errors.New(fmt.Sprintf("failed to get agent with UUID %s", agentUUID))
	}
	err := json.Unmarshal([]byte(agentInfoStr), &agentInfo)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to convert agent info, error=%s", err))
	}
	return &agentInfo, nil
}

func PutAgent(info AgentInstallInfo) (string, AgentInfo, error) {
	agentUUID := MakeAgentUUID(info)
	agentInfo := AgentInfo{}
	if strings.EqualFold(info.ServiceType, MCKS) || strings.EqualFold(info.ServiceType, MCKSAGENT_TYPE) || strings.EqualFold(info.ServiceType, MCKSAGENT_SHORTHAND_TYPE) {
		agentInfo = AgentInfo{
			NsId:                  info.NsId,
			McksID:                info.McksID,
			AgentType:             config.GetInstance().Monitoring.DefaultPolicy,
			AgentState:            string(Enable),
			AgentHealth:           string(Healthy),
			AgentUnhealthyRespCnt: 0,
			ServiceType:           info.ServiceType,
		}
	} else {
		agentInfo = AgentInfo{
			NsId:                  info.NsId,
			McisId:                info.McisId,
			VmId:                  info.VmId,
			CspType:               info.CspType,
			AgentType:             config.GetInstance().Monitoring.DefaultPolicy,
			AgentState:            string(Enable),
			AgentHealth:           string(Healthy),
			AgentUnhealthyRespCnt: 0,
			PublicIp:              info.PublicIp,
			ServiceType:           info.ServiceType,
		}
	}

	agentInfoBytes, err := json.Marshal(agentInfo)
	if err != nil {
		return "", AgentInfo{}, errors.New(fmt.Sprintf("failed to convert metadata format to json, error=%s", err))
	}
	err = cbstore.GetInstance().StorePut(types.Agent+agentUUID, string(agentInfoBytes))
	if err != nil {
		return "", AgentInfo{}, errors.New(fmt.Sprintf("failed to put metadata, error=%s", err))
	}
	return agentUUID, agentInfo, nil
}
