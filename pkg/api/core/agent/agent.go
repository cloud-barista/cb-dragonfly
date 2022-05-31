package agent

import (
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/common"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/mcis"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/mcks"
	"net/http"
	"strings"
)

func InstallAgent(info common.AgentInstallInfo) (int, error) {
	if agentMetadata, _ := common.GetAgent(info); agentMetadata != nil {
		return http.StatusBadRequest, errors.New(fmt.Sprintf("already exist agent, service_type: %s, namespace: %s", info.ServiceType, info.NsId))
	}

	mcksType := strings.EqualFold(info.ServiceType, common.MCKS) || strings.EqualFold(info.ServiceType, common.MCKSAGENT_TYPE) || strings.EqualFold(info.ServiceType, common.MCKSAGENT_SHORTHAND_TYPE)
	if mcksType {
		return mcks.InstallAgent(info)
	}
	return mcis.InstallAgent(info)
}

// 전체 에이전트 삭제 테스트용 코드
func UninstallAgent(info common.AgentInstallInfo) (int, error) {
	if agentMetadata, _ := common.GetAgent(info); agentMetadata == nil {
		return http.StatusBadRequest, errors.New(fmt.Sprintf("requested agent info not found, service_type: %s, namespace: %s", info.ServiceType, info.NsId))
	}
	mcksType := strings.EqualFold(info.ServiceType, common.MCKS) || strings.EqualFold(info.ServiceType, common.MCKSAGENT_TYPE) || strings.EqualFold(info.ServiceType, common.MCKSAGENT_SHORTHAND_TYPE)
	if mcksType {
		return mcks.UninstallAgent(info)
	}
	return mcis.UninstallAgent(info)
}
