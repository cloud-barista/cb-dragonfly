package metric

import (
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
)

// 멀티 클라우드 인프라 서비스 개별 VM 모니터링 정보 조회
func GetVMMonInfo(c echo.Context) error {
	// Path 파라미터 가져오기
	nsId := c.Param("ns")
	mcisId := c.Param("mcis_id")
	vmId := c.Param("vm_id")
	metricName := c.Param("metric_name")
	// Query 파라미터 가져오기
	period := c.QueryParam("periodType")
	aggregateType := c.QueryParam("statisticsCriteria")
	duration := c.QueryParam("duration")

	result, errCode, err := metric.GetVMMonInfo(nsId, mcisId, vmId, metricName, period, aggregateType, duration)
	if errCode != http.StatusOK {
		return echo.NewHTTPError(errCode, rest.SetMessage(err.Error()))
	}
	return c.JSON(http.StatusOK, result)
}

// 멀티 클라우드 인프라 서비스 개별 VM 최신 모니터링 정보 조회
//func GetVMLatestMonInfo(c echo.Context) error {
//	// Path 파라미터 가져오기
//	nsId := c.Param("ns")
//	mcisId := c.Param("mcis_id")
//	vmId := c.Param("vm_id")
//	metricName := c.Param("metric_name")
//	// Query 파라미터 가져오기
//	statisticsCriteria := c.QueryParam("statisticsCriteria")
//
//	result, errCode, err := metric.GetVMLatestMonInfo(nsId, mcisId, vmId, metricName, statisticsCriteria)
//	if err != nil {
//		return echo.NewHTTPError(errCode, rest.SetMessage(err.Error()))
//	}
//	return c.JSON(http.StatusOK, result)
//}
