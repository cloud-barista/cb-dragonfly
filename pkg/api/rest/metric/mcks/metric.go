package mcks

import (
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/metric"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/rest"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
)

// GetMCKSMonInfo 멀티 클라우드 쿠버네티스 서비스 모니터링 정보 조회
// @Summary Get Cluster monitoring info
// @Description 멀티 클라우드 쿠버네티스 서비스 모니터링 정보 조회
// @Tags [Monitoring] Monitoring management
// @Accept  json
// @Produce  json
// @Param ns_id path string true "네임스페이스 아이디"
// @Param mcks_id path string true "MCKS 아이디"
// @Param metric_name path string true "메트릭 정보"
// @Param periodType query string false "모니터링 단위" Enums(m, h, d)
// @Param statisticsCriteria query string false "모니터링 통계 기준" Enums(min, max, avg, last)
// @Param duration query string false "모니터링 조회 범위" Enums(5m, 5h, 5d)
// @Success 200 {object} rest.VMMonInfoType
// @Failure 404 {object} rest.SimpleMsg
// @Failure 500 {object} rest.SimpleMsg
// @Router /ns/{ns_id}/mcks/{mcks_id}/metric/{metric_name}/info [get]
func GetMCKSMonInfo(c echo.Context) error {
	// Path 파라미터 가져오기
	nsId := c.Param("ns_id")
	mcksId := c.Param("mcks_id")
	metricName := c.Param("metric_name")
	// Query 파라미터 가져오기
	period := c.QueryParam("periodType")
	aggregateType := c.QueryParam("statisticsCriteria")
	duration := c.QueryParam("duration")
	groupBy := c.QueryParam("groupBy")
	node := c.QueryParam("node")
	namespace := c.QueryParam("namespace")
	pod := c.QueryParam("pod")

	if strings.EqualFold(groupBy, types.Cluster) {
		if len(node) > 0 {
			if !strings.EqualFold(node, types.ALL) {
				return echo.NewHTTPError(http.StatusBadRequest, rest.SetMessage(fmt.Sprintf("monitoring for single node is not supported with groupBy, %s", groupBy)))
			}
		}
	}

	if strings.EqualFold(groupBy, types.Node) {
		if strings.EqualFold(node, types.ALL) || len(node) == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, rest.SetMessage("not supported node name for single node monitoring"))
		}
	}

	if strings.EqualFold(groupBy, string(types.MCKS_POD)) {
		if len(pod) > 0 {
			if len(namespace) == 0 {
				return echo.NewHTTPError(http.StatusBadRequest, rest.SetMessage("empty namespace parameter for pod monitoring"))
			}
		}
		if len(node) > 0 && len(namespace) > 0 {
			return echo.NewHTTPError(http.StatusBadRequest, rest.SetMessage("not supported monitoring"))
		}
	}

	if string(duration[len(duration)-1]) == "m" {
		durationInt, _ := strconv.Atoi(duration[:len(duration)-1])
		if durationInt < 2 {
			return echo.NewHTTPError(404, rest.SetMessage("Error! Mininum duration time is 2m"))
		}
	}

	dbInfo := types.DBMetricRequestInfo{
		NsID:                nsId,
		ServiceType:         types.MCKS,
		ServiceID:           mcksId,
		MetricName:          metricName,
		MonitoringMechanism: strings.EqualFold(config.GetInstance().Monitoring.DefaultPolicy, types.PushPolicy),
		Period:              period,
		MCKSReqInfo: types.MCKSReqInfo{
			GroupBy:   groupBy,
			Node:      node,
			Namespace: namespace,
			Pod:       pod,
		},
		AggegateType: aggregateType,
		Duration:     duration,
	}

	result, errCode, err := metric.GetMonInfo(dbInfo)
	if errCode != http.StatusOK {
		return echo.NewHTTPError(errCode, rest.SetMessage(err.Error()))
	}
	return c.JSON(http.StatusOK, result)
}
