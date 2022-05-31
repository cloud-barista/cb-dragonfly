package agent

import (
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/common"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"net/http"
	"strings"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest"
	"github.com/labstack/echo/v4"
)

type MetaDataListType struct {
	Id common.AgentInfo `json:"id(ns_id/mcis_id/vm_id/csp_type)"`
}

// ListAgentMetadata 에이전트 메타데이터 조회
// @Summary List agent metadata
// @Description 에이전트 메타데이터 조회
// @Tags [Agent] Monitoring Agent
// @Accept  json
// @Produce  json
// @Param ns query string false "네임스페이스 아이디" Enums(test_ns)
// @Param mcisId query string false "MCIS 아이디" Enums(test_mcis)
// @Param vmId query string false "VM 아이디" Enums(test_vm)
// @Param cspType query string false "VM의 CSP 정보" Enums(aws)
// @Success 200 {object}  rest.JSONResult{[DEFAULT]=[]MetaDataListType,[ID]=AgentInfo} "Different return structures by the given param"
// @Failure 404 {object} rest.SimpleMsg
// @Failure 500 {object} rest.SimpleMsg
// @Router /agent/metadata [get]
func ListAgentMetadata(c echo.Context) error {
	// 에이전트 UUID 파라미터 값 추출

	// 파라미터 값 체크
	agentMetadataList, err := common.ListAgent()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, rest.SetMessage(fmt.Sprintf("failed to get metadata list, error=%s", err)))
	}
	return c.JSON(http.StatusOK, agentMetadataList)
}

func GetAgentMetadata(c echo.Context) error {
	// 에이전트 UUID 파라미터 값 추출
	nsId := c.QueryParam("ns")
	serviceType := c.QueryParam("service_type")
	serviceId := c.QueryParam("service_id")
	var vmId, cspType string

	if !checkEmptyFormParam(serviceType) {
		return c.JSON(http.StatusBadRequest, rest.SetMessage("empty agent type parameter"))
	}

	requestInfo := common.AgentInstallInfo{
		ServiceType: serviceType,
		NsId:        nsId,
	}

	if strings.EqualFold(serviceType, common.MCKS) || strings.EqualFold(serviceType, common.MCKSAGENT_TYPE) || strings.EqualFold(serviceType, common.MCKSAGENT_SHORTHAND_TYPE) {
		if !checkEmptyFormParam(nsId, serviceId) {
			return c.JSON(http.StatusBadRequest, rest.SetMessage("bad request parameter to get mcks agent metadata"))
		}
		requestInfo.McksID = serviceId
	} else {
		vmId = c.QueryParam("vm_id")
		cspType = c.QueryParam("csp_type")
		if !checkEmptyFormParam(nsId, serviceId, vmId, cspType) {
			return c.JSON(http.StatusBadRequest, rest.SetMessage("bad request parameter to get mcis agent metadata"))
		}
		requestInfo.McksID = serviceId
		requestInfo.VmId = vmId
		requestInfo.CspType = cspType
	}

	agentMetadata, err := common.GetAgent(requestInfo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, rest.SetMessage(fmt.Sprintf("failed to get metadata, error=%s", err)))
	}
	return c.JSON(http.StatusOK, agentMetadata)
}

func PutAgentMetadata(c echo.Context) error {
	// 에이전트 UUID 파라미터 값 추출
	params := &rest.AgentType{}
	if err := c.Bind(params); err != nil {
		return err
	}

	if !checkEmptyFormParam(params.ServiceType) {
		return c.JSON(http.StatusBadRequest, rest.SetMessage("empty agent type parameter"))
	}

	if strings.EqualFold(params.ServiceType, common.MCKS) || strings.EqualFold(params.ServiceType, common.MCKSAGENT_TYPE) || strings.EqualFold(params.ServiceType, common.MCKSAGENT_SHORTHAND_TYPE) {
		// 토큰 값이 비어있을 경우
		if !checkEmptyFormParam(params.NsId, params.McksId) {
			return c.JSON(http.StatusBadRequest, rest.SetMessage("bad request parameter to update mcks agent metadata"))
		}
	}
	// MCIS 에이전트 form 파라미터 값 체크
	if strings.EqualFold(params.ServiceType, common.MCIS) || strings.EqualFold(params.ServiceType, common.MCISAGENT_TYPE) {
		// MCIS 에이전트 form 파라미터 값 체크
		if !checkEmptyFormParam(params.NsId, params.McisId, params.VmId, params.CspType, params.PublicIp) {
			return c.JSON(http.StatusBadRequest, rest.SetMessage("bad request parameter to update mcis agent metadata"))
		}
	}

	requestInfo := common.AgentInstallInfo{
		ServiceType: params.ServiceType,
		NsId:        params.NsId,
		McisId:      params.McisId,
		VmId:        params.VmId,
		PublicIp:    params.PublicIp,
		CspType:     params.CspType,
		McksID:      params.McksId,
	}
	// 메타데이터 조회
	if _, err := common.GetAgent(requestInfo); err != nil {
		return c.JSON(http.StatusInternalServerError, rest.SetMessage(fmt.Sprintf("failed to get metadata before update metadata, error=%s", err)))
	}
	// 메타데이터 수정
	agentUUID, agentMetadata, err := common.PutAgent(requestInfo)
	errQue := util.RingQueuePut(types.TopicAdd, agentUUID)
	if err != nil || errQue != nil {
		return c.JSON(http.StatusInternalServerError, rest.SetMessage(fmt.Sprintf("failed to update metadata, error=%s", err)))
	}
	return c.JSON(http.StatusOK, agentMetadata)
}
