package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/collector"
	"github.com/labstack/echo/v4"
)

// 멀티클라우드 인프라 VM 온디멘드 모니터링
func OndemandAllMetric(c echo.Context) error {
	//온디멘드 모니터링 파라미터 추출
	publicIP := c.Param("public_ip")

	// Query 파라미터 값 체크
	if publicIP == "" {
		err := errors.New("failed to get package. query parameter is missing")
		return c.JSON(http.StatusInternalServerError, err)
	}
	resp, err := http.Get(fmt.Sprintf("http://%s:8080/cb-dragonfly/metric", publicIP))
	if err != nil {
		return c.String(http.StatusNotImplemented, "Server Closed")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	var data = map[string]collector.TelegrafMetric{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, data)
}

// 멀티클라우드 인프라 VM 온디멘드 모니터링
func OndemandMetric(c echo.Context) error {
	//온디멘드 모니터링 Agent IP 파라미터 추출
	publicIP := c.Param("public_ip")

	// Query Agent IP 값 체크
	if publicIP == "" {
		err := errors.New("No Agent IP in API")
		return c.JSON(http.StatusInternalServerError, err)
	}
	//온디멘드 모니터링 매트릭 파라미터 추출
	metrictype := c.Param("metrictype")

	//Query 매트릭 값 체크
	if metrictype == "" {
		err := errors.New("No Metric Type in API")
		return c.JSON(http.StatusInternalServerError, err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:8080/cb-dragonfly/metric/%s", publicIP, metrictype))
	if err != nil {
		return c.String(http.StatusNotImplemented, "Server Closed")
	}
	defer resp.Body.Close()
	var data = map[string]collector.TelegrafMetric{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, data)
}
