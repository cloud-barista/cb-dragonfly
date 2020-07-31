package metric

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/metric"
)

// 멀티 클라우드 인프라 서비스 모니터링 정보 조회
func GetMCISMonInfo(c echo.Context) error {
	nsId := c.Param("ns")
	mcisId := c.Param("mcis_id")

	result, err := metric.GetMCISMonInfo(nsId, mcisId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// 멀티 클라우드 인프라 서비스(MCIS) 실시간 모니터링 정보 조회
func GetMCISRealtimeMonInfo(c echo.Context) error {
	nsId := c.Param("ns")
	mcisId := c.Param("mcis_id")

	result, err := metric.GetMCISRealtimeMonInfo(nsId, mcisId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}
