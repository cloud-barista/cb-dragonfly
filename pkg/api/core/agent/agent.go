package agent

import (
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/common"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/mcis"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/mck8s"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
)

func InstallAgent(info common.AgentInstallInfo) (int, error) {
	switch config.GetInstance().Monitoring.DeployType {
	case types.Dev, types.Compose:
		if agentMetadata, _ := common.GetAgent(info); agentMetadata != nil {
			return http.StatusBadRequest, errors.New(fmt.Sprintf("already exist agent, service_type: %s, namespace: %s", info.ServiceType, info.NsId))
		}
	}

	if util.CheckMCK8SType(info.ServiceType) {
		_, domain, _, err := util.GetProtocolDomainPort(info.APIServerURL)
		if err != nil {
			return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to get domain info from request, error=%s", err.Error()))
		}
		if err = mck8s.HandleDomain(info.PrivateDomain, types.CREATE, *info.IP, domain); err != nil {
			return http.StatusInternalServerError, err
		}

		status, err := mck8s.InstallAgent(info)
		if err != nil {
			_ = mck8s.HandleDomain(info.PrivateDomain, types.DELETE, *info.IP, domain)
		}
		return status, nil
	}
	return mcis.InstallAgent(info)
}

// UninstallAgent 전체 에이전트 삭제 테스트용 코드
func UninstallAgent(info common.AgentInstallInfo) (int, error) {
	switch config.GetInstance().Monitoring.DeployType {
	case types.Dev, types.Compose:
		if agentMetadata, _ := common.GetAgent(info); agentMetadata == nil {
			return http.StatusBadRequest, errors.New(fmt.Sprintf("requested agent info not found, service_type: %s, namespace: %s", info.ServiceType, info.NsId))
		}
	}

	if util.CheckMCK8SType(info.ServiceType) {
		_, domain, _, err := util.GetProtocolDomainPort(info.APIServerURL)
		if err != nil {
			return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to get domain info from request, error=%s", err.Error()))
		}
		if err = mck8s.HandleDomain(info.PrivateDomain, types.DELETE, *info.IP, domain); err != nil {
			return http.StatusInternalServerError, err
		}
		status, err := mck8s.UninstallAgent(info)
		if err != nil {
			_ = mck8s.HandleDomain(info.PrivateDomain, types.CREATE, *info.IP, domain)
		}
		return status, nil
	}
	return mcis.UninstallAgent(info)
}
