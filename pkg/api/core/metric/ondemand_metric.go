package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent"
	"io/ioutil"
	"net/http"
	"sync"
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

type PacketsInfo struct {
	DestinationIp string
	PacketCnt int
	TotalPacketBytes int
	Msg string
}

type NetworkPacketsResult struct {
	WatchTime string
	PacketsInfos map[int] PacketsInfo
}

func GetMCISOnDemandPacketInfo(nsId string, mcisId string, vmId string, watchTime string) (NetworkPacketsResult, int, error) {
	agentList, err := agent.ListAgent()
	if err != nil {
		fmt.Println("Fail to Get AgentList From CB-Store")
		return NetworkPacketsResult{}, http.StatusInternalServerError, err
	}
	var sourceAgentIP string
	var targetAgentInfo []agent.AgentInfo

	for _, agentMetadata := range agentList {
		if agentMetadata.McisId == mcisId  && agentMetadata.NsId == nsId {
			if agentMetadata.VmId == vmId {
				sourceAgentIP = agentMetadata.PublicIp
			} else {
				targetAgentInfo = append(targetAgentInfo, agentMetadata)
			}
		}
	}

	if sourceAgentIP == "" || len(targetAgentInfo) == 0 {
		return NetworkPacketsResult{}, http.StatusOK, nil
	}

	client := http.Client{
		Timeout: AgentTimeout * time.Second,
	}

	wg := sync.WaitGroup{}
	wg.Add(len(targetAgentInfo))

	result := NetworkPacketsResult{
		WatchTime: watchTime,
		PacketsInfos: map[int] PacketsInfo{},
	}

	for idx, targetAgent := range targetAgentInfo {
		agentUrl := fmt.Sprintf("http://%s:%d/cb-dragonfly/mcis/dstip/%s/watchtime/%s", sourceAgentIP, AgentPort, targetAgent.PublicIp, watchTime)
		idx := idx
		targetAgent := targetAgent
		go func() {
			defer wg.Done()
			packetsInfo := PacketsInfo{}
			resp, err := client.Get(agentUrl)
			if err != nil {
				fmt.Println("err: " + targetAgent.PublicIp+", msg: ", err)
				return
			}
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				fmt.Println("err: " + targetAgent.PublicIp+", msg: ", err2)
				return
			}
			_ = json.Unmarshal(body, &packetsInfo)
			result.PacketsInfos[idx] = packetsInfo
		}()
	}

	return result, http.StatusOK, err
}

type ProcessUsage struct {
	Pid string
	CpuUsage string
	MemUsage string
	Command string
}

func GetMCISOnDemandProcessInfo(publicIp string) (map[string] []ProcessUsage, int, error) {

	userProcess := map[string] []ProcessUsage{}
	client := http.Client{
		Timeout: AgentTimeout * time.Second,
	}
	agentUrl := fmt.Sprintf("http://%s:%d/cb-dragonfly/mcis/process", publicIp, AgentPort)
	resp, err := client.Get(agentUrl)
	if err != nil {
		return map[string] []ProcessUsage{}, http.StatusInternalServerError, err
	}
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return map[string] []ProcessUsage{}, http.StatusInternalServerError, err
	}
	_ = json.Unmarshal(body, &userProcess)

	return userProcess, http.StatusOK, nil

}

type MCISGroupSpecs struct {
	NetBwGbps int
	NumCore int
	NumGpu int
	NumStorage int
	NumvCPU int
	StorageGiB int
}

func GetMCISSpecInfo(nsId string, mcisId string) (MCISGroupSpecs, int, error) {

	mcisGroupSpecs := MCISGroupSpecs{}
	//agentMetadataList, err := agent.ListAgent()
	//if err != nil {
	//	return mcisGroupSpecs, http.StatusInternalServerError, err
	//}

	//for _, agentMetadata := range agentMetadataList {
		//getVmSpecIdUrl := fmt.Sprintf("")
		//agentMetadata.VmId
	//}

	return mcisGroupSpecs, http.StatusOK, nil
}
