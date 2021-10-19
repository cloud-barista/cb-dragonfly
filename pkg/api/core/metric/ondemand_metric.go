package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/modules/procedure/push/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
)

const (
	AgentPort    = 8888
	AgentTimeout = 10
)

func GetVMOnDemandMonInfo(metricName string, publicIP string) (interface{}, int, error) {
	metric := types.Metric(metricName)

	// 메트릭 타입 유효성 체크
	if metric == types.None {
		return nil, http.StatusInternalServerError, errors.New(fmt.Sprintf("not found metric : %s", metricName))
	}

	// disk, diskio 메트릭 조회
	if metric == types.Disk {
		diskMetric, err := getVMOnDemandMonInfo(types.Disk, publicIP)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		diskioMetric, err := getVMOnDemandMonInfo(types.DiskIO, publicIP)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		diskMetricMap := diskMetric["values"].(map[string]interface{})
		diskioMetricMap := diskioMetric["values"].(map[string]interface{})
		for k, v := range diskioMetricMap {
			diskMetricMap[k] = v
		}

		return diskMetric, http.StatusOK, nil
	}

	// cpu, cpufreq, memory, network 메트릭 조회
	resultMetric, err := getVMOnDemandMonInfo(metric, publicIP)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return resultMetric, http.StatusOK, nil
}

func getVMOnDemandMonInfo(metric types.Metric, publicIP string) (map[string]interface{}, error) {
	client := http.Client{
		Timeout: AgentTimeout * time.Second,
	}
	agentUrl := fmt.Sprintf("http://%s:%d/cb-dragonfly/metric/%s", publicIP, AgentPort, metric.ToAgentMetricKey())
	resp, err := client.Get(agentUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var metricData = map[string]collector.TelegrafMetric{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &metricData)
	if err != nil {
		return nil, err
	}
	resultMetric, err := collector.ConvertMonMetric(metric, metricData[metric.ToAgentMetricKey()])
	if err != nil {
		return nil, err
	}
	return resultMetric, nil
}
