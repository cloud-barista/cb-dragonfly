package puller

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
	"github.com/cloud-barista/cb-dragonfly/pkg/metadata"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdb/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
)

type PullCaller struct {
	AgentList map[string]metadata.AgentInfo
}

func NewPullCaller(agentList map[string]metadata.AgentInfo) (PullCaller, error) {
	return PullCaller{AgentList: agentList}, nil
}

func (pc PullCaller) StartPull() {
	for key, agent := range pc.AgentList {
		fmt.Println("PULL AGENT : " + key)
		go pc.pullMetric(key, agent.PublicIp)
	}
}

func (pc PullCaller) pullMetric(uuid string, ip string) {
	metricArr := []types.MetricType{types.CPU, types.CPUFREQ, types.DISK, types.DISKIO, types.NETWORK}
	for _, pullMetric := range metricArr {
		metricKey := string(pullMetric)
		fmt.Printf("[%s] CALL API: http://%s:8888/%s\n", uuid, ip, string(metricKey))

		//TODO: call API for pull
		result, errCode, err := metric.GetVMOnDemandMonInfo("", "", "", metricKey, ip)
		fmt.Println(result, errCode, err)
		if errCode != http.StatusOK {
			fmt.Println(err)
			continue
			//TODO: Update Agent Health Status to Unhealthy
		}
		if result == nil {
			continue
		}

		// 메트릭 정보 파싱
		metricData := result.(map[string]interface{})
		metricName := metricData["name"].(string)
		tagArr := map[string]string{}
		for k, v := range metricData["tags"].(map[string]interface{}) {
			tagArr[k] = v.(string)
		}
		metricVal := metricData["values"].(map[string]interface{})

		// 메트릭 정보 InfluxDB 저장
		err = influxdbv1.GetInstance().WriteOnDemandMetric(metricName, tagArr, metricVal)
		if err != nil {
			fmt.Println(err)
		}
	}
}
