package agent

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent"
)

func InstallTelegraf(c echo.Context) error {
	// form 파라미터 값 가져오기
	nsId := c.FormValue("ns_id")
	mcisId := c.FormValue("mcis_id")
	vmId := c.FormValue("vm_id")
	publicIp := c.FormValue("public_ip")
	userName := c.FormValue("user_name")
	sshKey := c.FormValue("ssh_key")

	// form 파라미터 값 체크
	if nsId == "" || mcisId == "" || vmId == "" || publicIp == "" || userName == "" || sshKey == "" {
		errMsg := setMessage("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, errMsg)
	}

	err := agent.InstallTelegraf(nsId, mcisId, vmId, publicIp, userName, sshKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	successMsg := setMessage("agent installation is finished")
	return c.JSON(http.StatusOK, successMsg)
}

func setMessage(msg string) echo.Map {
	errResp := echo.Map{}
	errResp["message"] = msg
	return errResp
}
