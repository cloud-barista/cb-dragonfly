package metric

import (
	"errors"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest"
	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
	"github.com/labstack/echo/v4"
	"net/http"
)

// 멀티 클라우드 인프라 서비스 모니터링 정보 조회
func GetMCISMonInfo(c echo.Context) error {
	nsId := c.Param("ns")
	mcisId := c.Param("mcis_id")

	result, err := metric.GetMCISMonInfo(nsId, mcisId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, rest.SetMessage(err.Error()))
	}
	return c.JSON(http.StatusOK, result)
}

// 멀티 클라우드 인프라 서비스(MCIS) 실시간 모니터링 정보 조회
func GetMCISRealtimeMonInfo(c echo.Context) error {
	nsId := c.Param("ns")
	mcisId := c.Param("mcis_id")

	result, err := metric.GetMCISRealtimeMonInfo(nsId, mcisId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, rest.SetMessage(err.Error()))
	}
	return c.JSON(http.StatusOK, result)
}

// 멀티 클라우드 인트라 서비스 모니터링 메트릭 수집
func GetMCISMetric(c echo.Context) error {
	var mcismetric metric.MCISMetric

	// 메트릭 확인
	metrictype := c.Param("mcis_metric_name")
	if metrictype == "" {
		err := errors.New("No Metric Type in API")
		return c.JSON(http.StatusInternalServerError, err)
	}
	// MCIS 모니터링 메트릭 파라미터 기반 동작
	switch metrictype {
	case "Rtt":
		return mcismetric.Rtt(c)
	case "Mrtt":
		return mcismetric.Mrtt(c)
	default:
		return mcismetric.MCISMetric(c)
	}
	return nil
}
