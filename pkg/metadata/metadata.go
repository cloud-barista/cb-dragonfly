package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest"
	"github.com/cloud-barista/cb-dragonfly/pkg/cbstore"
	"github.com/labstack/echo/v4"
)

const (
	AGENT_LIST = "agentlist"
	ENABLE     = "Enable"
	DISABLE    = "Disable"
)

// AgentHealth 에이전트 API 헬스체크 상태
type AgentHealth string

const (
	HEALTHY   AgentHealth = "healthy"
	UNHEALTHY AgentHealth = "unhealthy"
)

//type Metadata struct {
//	Key   string
//	Value *AgentInfo
//}

// AgentInfo 에이전트 상세 정보
type AgentInfo struct {
	AgentState  string `json:"agent_state"`
	AgentType   string `json:"agent_type"`
	PublicIp    string `json:"public_ip"`
	AgentHealth string `json:"agent_health"`
}

//var metadata = &Metadata{}
//var agentInfo = &AgentInfo{}

/*func setAgentInfo(agentState string, agentType string, publicIp string, agentHealth string) {
	getAgentInfo().AgentState = agentState
	getAgentInfo().AgentType = agentType
	getAgentInfo().PublicIp = publicIp
	getAgentInfo().AgentHealth = agentHealth
}
func getAgentInfo() *AgentInfo {
	return agentInfo
}*/

/*
func setMetadata(uuid string, agentInfo *AgentInfo) {
	getMetadata().Key = uuid
	getMetadata().Value = agentInfo
}
func getMetadata() *Metadata {
	return metadata
}
*/

func newAgentInfo(publicIp string) AgentInfo {
	return AgentInfo{
		AgentState:  ENABLE,
		AgentType:   ENABLE,
		PublicIp:    publicIp,
		AgentHealth: string(HEALTHY),
	}
}

// AgentListManager 에이전트 목록 관리
type AgentListManager struct{}

func (a AgentListManager) getAgentListFromStore() (map[string]AgentInfo, error) {
	agentList := map[string]AgentInfo{}
	agentListStr := cbstore.GetInstance().StoreGet(AGENT_LIST)
	if agentListStr != "" {
		err := json.Unmarshal([]byte(agentListStr), &agentList)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to convert agent list, error=%s", err))
		}
	}
	return agentList, nil
}

func (a AgentListManager) putAgentListToStore(agentList map[string]AgentInfo) error {
	agentListBytes, err := json.Marshal(agentList)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to convert agentList format to json, error=%s", err))
	}
	err = cbstore.GetInstance().Store.Put(AGENT_LIST, string(agentListBytes))
	if err != nil {
		return errors.New(fmt.Sprintf("failed to put agentList, error=%s", err))
	}
	return nil
}

func (a AgentListManager) AddAgent(uuid string, agentInfo AgentInfo) error {
	agentList, err := a.getAgentListFromStore()
	if err != nil {
		return err
	}

	if _, ok := agentList[uuid]; ok {
		return errors.New(fmt.Sprintf("failed to add agent, agent with UUID %s already exist", uuid))
	}
	agentList[uuid] = agentInfo

	return a.putAgentListToStore(agentList)
}

func (a AgentListManager) UpdateAgent(uuid string, agentInfo AgentInfo) error {
	agentList, err := a.getAgentListFromStore()
	if err != nil {
		return err
	}

	if _, ok := agentList[uuid]; !ok {
		return errors.New(fmt.Sprintf("failed to update agent, agent with UUID %s not exist", uuid))
	}
	agentList[uuid] = agentInfo

	return a.putAgentListToStore(agentList)
}

func (a AgentListManager) DeleteAgent(uuid string) error {
	agentList, err := a.getAgentListFromStore()
	if err != nil {
		return err
	}

	if _, ok := agentList[uuid]; !ok {
		return errors.New(fmt.Sprintf("failed to update agent, agent with UUID %s not exist", uuid))
	}
	delete(agentList, uuid)

	return a.putAgentListToStore(agentList)
}

func (a AgentListManager) GetAgentList() (map[string]AgentInfo, error) {
	return a.getAgentListFromStore()
}

func (a AgentListManager) GetAgentInfo(uuid string) (AgentInfo, error) {
	agentInfo := AgentInfo{}
	agentInfoStr := cbstore.GetInstance().StoreGet(uuid)

	if agentInfoStr != "" {
		err := json.Unmarshal([]byte(agentInfoStr), &agentInfo)
		if err != nil {
			return AgentInfo{}, errors.New(fmt.Sprintf("failed to convert agent info, error=%s", err))
		}
	}
	return agentInfo, nil
}

func AgentInstallationMetadata(nsId string, mcisId string, vmId string, cspType string, publicIp string) error {
	//setAgentInfo(agentState, agentType, publicIp, string(agentHealth))
	//setMetadata(makeAgentUUID(nsId, mcisId, vmId, cspType), getAgentInfo())

	agentUUID := makeAgentUUID(nsId, mcisId, vmId, cspType)
	agentInfo := newAgentInfo(publicIp)

	// 에이전트 목록 추가
	var agentList AgentListManager
	err := agentList.AddAgent(agentUUID, agentInfo)
	if err != nil {
		return err
	}

	// 에이전트 메타데이터 등록
	agentInfoBytes, err := json.Marshal(agentInfo)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to convert metadata format to json, error=%s", err))
	}
	err = cbstore.GetInstance().StorePut(agentUUID, string(agentInfoBytes))
	if err != nil {
		return errors.New(fmt.Sprintf("failed to put metadata, error=%s", err))
	}
	return nil
}

func AgentDeletionMetadata(nsId string, mcisId string, vmId string, cspType string, publicIp string) error {
	//agent_State := DISABLE
	//agent_Type := DISABLE
	//setAgentInfo(&agent_State, &agent_Type, &publicIp)
	//setMetadata(makeAgentUUID(nsId, mcisId, vmId, cspType), getAgentInfo())

	//_, err := json.Marshal(getMetadata().Value)
	//if err != nil {
	//	return errors.New(fmt.Sprintf("failed to convert metadata format to json, error=%s", err))
	//}

	agentUUID := makeAgentUUID(nsId, mcisId, vmId, cspType)

	// 에이전트 목록 삭제
	var agentList AgentListManager
	err := agentList.DeleteAgent(agentUUID)
	if err != nil {
		return err
	}

	// 에이전트 메타데이터 삭제
	err = cbstore.GetInstance().StoreDelete(agentUUID)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to delete metadata, error=%s", err))
	}
	return nil
}

func makeAgentUUID(nsId string, mcisId string, vmId string, cspType string) string {
	UUID := fmt.Sprintf(nsId + "/" + mcisId + "/" + vmId + "/" + cspType)
	return UUID
}

func ShowMetadata(c echo.Context) error {
	// 에이전트 UUID 파라미터 값 추출
	nsId := c.Param("ns")
	mcisId := c.Param("mcis_id")
	vmId := c.Param("vm_id")
	cspType := c.Param("csp_type")

	// 파라미터 값 체크
	if nsId == "" || mcisId == "" || vmId == "" || cspType == "" {
		return c.JSON(http.StatusInternalServerError, rest.SetMessage("failed to get metadata. parameter is missing."))
	}
	metadata, err := cbstore.GetInstance().Store.Get(fmt.Sprintf(nsId + "/" + mcisId + "/" + vmId + "/" + cspType))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.New(fmt.Sprintf("Get Data from CB-Store Error, err=%s", err)))
	}

	return c.JSON(http.StatusOK, metadata)
}
