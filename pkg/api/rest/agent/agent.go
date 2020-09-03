package agent

import (
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/agent"
)

func InstallTelegraf(c echo.Context) error {
	// form 파라미터 값 가져오기
	nsId := c.FormValue("ns_id")
	mcisId := c.FormValue("mcis_id")
	vmId := c.FormValue("vm_id")
	publicIp := c.FormValue("public_ip")
	userName := c.FormValue("user_name")
	sshKey := c.FormValue("ssh_key")
	cspType := c.FormValue("cspType")

	// form 파라미터 값 체크
	if nsId == "" || mcisId == "" || vmId == "" || publicIp == "" || userName == "" || sshKey == "" || cspType == "" {
		return c.JSON(http.StatusInternalServerError, rest.SetMessage("failed to get package. query parameter is missing"))
	}

	errCode, err := agent.InstallTelegraf(nsId, mcisId, vmId, publicIp, userName, sshKey, cspType)
	if errCode != http.StatusOK {
		return c.JSON(errCode, rest.SetMessage(err.Error()))
	}
	return c.JSON(http.StatusOK, rest.SetMessage("agent installation is finished"))
}
