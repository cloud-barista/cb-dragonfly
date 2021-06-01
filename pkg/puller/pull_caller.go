package puller

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
	"github.com/cloud-barista/cb-dragonfly/pkg/metadata"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdb/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
)

const (
	AgentUnhealthyCnt = 5
)

type PullCaller struct {
	AgentList map[string]metadata.AgentInfo
}

func NewPullCaller(agentList map[string]metadata.AgentInfo) (PullCaller, error) {
	return PullCaller{AgentList: agentList}, nil
}

func (pc PullCaller) StartPull() {
	var wg sync.WaitGroup
	for key, agent := range pc.AgentList {
		// Check agent status
		if agent.AgentState == string(metadata.Disable) {
			continue
		}
		// Check agent health
		if agent.AgentHealth == string(metadata.Unhealthy) {
			// TODO: Call healthcheck API
			continue
		}
		wg.Add(1)
		go pc.pullMetric(&wg, key, agent)
	}
	wg.Wait()
}

func (pc PullCaller) pullMetric(wg *sync.WaitGroup, uuid string, agentInfo metadata.AgentInfo) {
	defer wg.Done()
	metricArr := []types.Metric{types.Cpu, types.CpuFrequency, types.Memory, types.Disk, types.Network}
	for _, pullMetric := range metricArr {

		if agentInfo.AgentState == string(metadata.Disable) || agentInfo.AgentHealth == string(metadata.Unhealthy) {
			// TODO: Call healthcheck API
			continue
		}

		fmt.Printf("[%s] CALL API: http://%s:%d/cb-dragonfly/metric/%s\n", uuid, agentInfo.PublicIp, metric.AgentPort, pullMetric.ToAgentMetricKey())

		// Pulling agent
		result, statusCode, err := metric.GetVMOnDemandMonInfo(pullMetric.ToString(), agentInfo.PublicIp)

		// Update Agent Health
		updated := false
		if statusCode == http.StatusOK && agentInfo.AgentHealth == string(metadata.Unhealthy) {
			updated = true
			agentInfo.AgentHealth = string(metadata.Healthy)
		}
		if statusCode != http.StatusOK {
			updated = true
			agentInfo.AgentUnhealthyRespCnt += 1
			if agentInfo.AgentUnhealthyRespCnt > AgentUnhealthyCnt {
				agentInfo.AgentHealth = string(metadata.Unhealthy)
			}
		}

		if updated {
			err := metadata.PutAgentMetadataToStore(uuid, agentInfo)
			if err != nil {
				continue
			}
			test := metadata.AgentListManager{}
			list, _ := test.GetAgentList()
			fmt.Println("===================================================")
			fmt.Println(list)
			fmt.Println("===================================================")
		}

		if result == nil {
			continue
		}

		// 메트릭 정보 파싱
		metricData := result.(map[string]interface{})
		metricName := metricData["name"].(string)
		if metricName == "" {
			continue
		}
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
