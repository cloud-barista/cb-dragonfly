package config

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/config"
)

// 모니터링 정책 설정
func SetMonConfig(c echo.Context) error {
	// form 파라미터 정보 가져오기
	agentInterval, err := strconv.Atoi(c.FormValue("agent_interval"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	collectorInterval, err := strconv.Atoi(c.FormValue("collector_interval"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	schedulingInterval, err := strconv.Atoi(c.FormValue("schedule_interval"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	maxHostCnt, err := strconv.Atoi(c.FormValue("max_host_count"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	agentTtl, err := strconv.Atoi(c.FormValue("agent_TTL"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	monConfig, errCode, err := config.SetMonConfig(agentInterval, collectorInterval, schedulingInterval, maxHostCnt, agentTtl)
	if errCode != http.StatusOK {
		return echo.NewHTTPError(errCode, err.Error())
	}
	return c.JSON(http.StatusOK, monConfig)
}

// 모니터링 정책 조회
func GetMonConfig(c echo.Context) error {
	monConfig, errCode, err := config.GetMonConfig()
	if errCode != http.StatusOK {
		return echo.NewHTTPError(errCode, err.Error())
	}
	return c.JSON(http.StatusOK, monConfig)
}

// 모니터링 정책 초기화
func ResetMonConfig(c echo.Context) error {
	monConfig, errCode, err := config.ResetMonConfig()
	if errCode != http.StatusOK {
		return echo.NewHTTPError(errCode, err.Error())
	}
	return c.JSON(http.StatusOK, monConfig)
}
