package metric

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/metric"
)

// 멀티 클라우드 인프라 VM 모니터링 정보 조회
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
		return echo.NewHTTPError(errCode, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// 멀티 클라우드 인프라 VM 실시간 모니터링 정보 조회
func GetVMRealtimeMonInfo(c echo.Context) error {
	// Path 파라미터 가져오기
	nsId := c.Param("ns")
	mcisId := c.Param("mcis_id")
	vmId := c.Param("vm_id")
	metricName := c.Param("metric_name")
	// Query 파라미터 가져오기
	statisticsCriteria := c.QueryParam("statisticsCriteria")

	result, err := metric.GetVMRealtimeMonInfo(nsId, mcisId, vmId, metricName, statisticsCriteria)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

type Config struct {
	props string
}

var config *Config
var once sync.Once

func GetInstance() *Config {
	once.Do(func() {
		// Initialize config struct
		config = &Config{
			props: "sample-data",
		}
	})
	// return config instance
	return config
}
