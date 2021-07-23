package agent

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest"
	"github.com/cloud-barista/cb-dragonfly/pkg/core/agent"
	"github.com/labstack/echo/v4"
)

var agentListManager agent.AgentListManager

// JSONResult's data field will be overridden by the specific type
type JSONResult struct {
	//Code    int          `json:"code" `
	//Message string       `json:"message"`
	//Data    interface{}  `json:"data"`
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
// @Success 200 {object}  JSONResult{[DEFAULT]=[]rest.AgentMetaDataListType,[ID]=AgentInfo} "Different return structures by the given param"
// @Failure 404 {object} rest.SimpleMsg
// @Failure 500 {object} rest.SimpleMsg
// @Router /agent/metadata [get]
func ListAgentMetadata(c echo.Context) error {
	// 에이전트 UUID 파라미터 값 추출
	nsId := c.QueryParam("ns")
	mcisId := c.QueryParam("mcisId")
	vmId := c.QueryParam("vmId")
	cspType := c.QueryParam("cspType")

	// 파라미터 값 체크
	if nsId == "" || mcisId == "" || vmId == "" || cspType == "" {
		agentMetadataList, err := agentListManager.GetAgentList()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, rest.SetMessage(fmt.Sprintf("failed to get metadata list, error=%s", err)))
		}
		return c.JSON(http.StatusOK, agentMetadataList)
	} else {
		agentUUID := agent.MakeAgentUUID(nsId, mcisId, vmId, cspType)
		agentMetadata, err := agentListManager.GetAgentInfo(agentUUID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, rest.SetMessage(fmt.Sprintf("failed to get metadata, error=%s", err)))
		}
		return c.JSON(http.StatusOK, agentMetadata)
	}
}
