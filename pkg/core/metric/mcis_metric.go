package metric

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

type MCISMetric struct {
	client    http.Client
	data      CBMCISMetric
	mdata     MCBMCISMetric
	mrequest  Mrequest
	request   Request
	parameter Parameter
}

func GetMCISMonInfo(nsId string, mcisId string) (interface{}, error) {
	// TODO: MCIS 서비스 모니터링 정보 조회 기능 개발
	return nil, nil
}

func GetMCISRealtimeMonInfo(nsId string, mcisId string) (interface{}, error) {
	// TODO: MCIS 서비스 실시간 모니터링 정보 조회 기능 개발
	return nil, nil
}

// MCISMetric ...
func (mc *MCISMetric) MCISMetric(c echo.Context) error {
	// API 기반 필요 파라미터 추출
	_ = mc.CheckParameter(c)

	// MCIS Get 요청 API 생성
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8080/cb-dragonfly/mcis/metric/%s", mc.parameter.agent_ip, mc.parameter.mcis_metric), nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	resp, err := mc.client.Do(req)
	if err != nil {
		return c.JSON(http.StatusNotImplemented, "Server Closed")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	err = json.Unmarshal(body, &mc.data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, &mc.data)
}

// Rtt ...
func (mc *MCISMetric) Rtt(c echo.Context) error {
	// API Body 데이터 추출
	if err := c.Bind(&mc.request); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	// API 기반 필요 파라미터 추출
	_ = mc.CheckParameter(c)

	// MCIS Get 요청 API 생성
	payload, err := json.Marshal(mc.request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8080/cb-dragonfly/mcis/metric/%s", mc.parameter.agent_ip, mc.parameter.mcis_metric), bytes.NewBuffer(payload))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := mc.client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = json.Unmarshal(body, &mc.data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, &mc.data)

}

// Mrtt ...
func (mc *MCISMetric) Mrtt(c echo.Context) error {
	// API Body 데이터 추출
	if err := c.Bind(&mc.mrequest); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	// API 기반 필요 파라미터 추출
	_ = mc.CheckParameter(c)

	// MCIS Get 요청 API 생성
	payload, err := json.Marshal(mc.mrequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:8080/cb-dragonfly/mcis/metric/%s", mc.parameter.agent_ip, mc.parameter.mcis_metric), bytes.NewBuffer(payload))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := mc.client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	err = json.Unmarshal(body, &mc.mdata)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, &mc.mdata)
}

func (mc *MCISMetric) CheckParameter(c echo.Context) error {
	mc.parameter.agent_ip = c.Param("agent_ip")

	// Query Agent IP 값 체크
	if mc.parameter.agent_ip == "" {
		err := errors.New("No Agent IP in API")
		return c.JSON(http.StatusInternalServerError, err)
	}
	// MCIS 모니터링 메트릭 파라미터 추출
	mc.parameter.mcis_metric = c.Param("mcis_metric_name")
	if mc.parameter.mcis_metric == "" {
		err := errors.New("No Metric Type in API")
		return c.JSON(http.StatusInternalServerError, err)
	}
	return nil
}
