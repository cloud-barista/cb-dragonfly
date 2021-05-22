package puller

import (
	"fmt"
	"net/http"

	"github.com/cloud-barista/cb-dragonfly/pkg/core/metric"
	"github.com/cloud-barista/cb-dragonfly/pkg/metadata"
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
		go pc.pullMetric(key, agent.PublicIp, agent.Port)
	}
}

func (pc PullCaller) pullMetric(uuid string, ip string, port string) {
	metricArr := []types.MetricType{types.CPU, types.CPUFREQ, types.DISK, types.DISKIO, types.NETWORK}
	for _, pullMetric := range metricArr {
		metricKey := string(pullMetric)
		fmt.Printf("[%s] CALL API: http://%s:%s/%s\n", uuid, ip, port, string(metricKey))

		// TODO: call API for pull
		result, errCode, err := metric.GetVMOnDemandMonInfo("", "", "", metricKey, ip)
		fmt.Println(result, errCode, err)
		if errCode != http.StatusOK {
			fmt.Println(err)
			// TODO: Update Agent Health Status to Unhealthy
		}
	}
}
