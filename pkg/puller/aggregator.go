package puller

import (
	"encoding/json"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/cbstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/metadata"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdb"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdb/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/sirupsen/logrus"
	"time"
)

type PullAggregator struct {
	AgentListManager metadata.AgentListManager
	Storage          influxdbv1.Storage
	CBStore          cbstore.CBStore
	AgentList        map[string]metadata.AgentInfo
}

func NewPullAggregator() (*PullAggregator, error) {
	pullAggregator := PullAggregator{
		AgentListManager: metadata.AgentListManager{},
		Storage:          *influxdbv1.GetInstance(),
		CBStore:          *cbstore.GetInstance(),
	}
	return &pullAggregator, nil
}

func (pa *PullAggregator) StartAggregate() error {
	metricArr := []types.Metric{types.Cpu, types.CpuFrequency, types.Memory, types.Disk, types.Network, types.DiskIO}
	aggregateInterval := time.Duration(config.GetInstance().Monitoring.PullerAggregateInterval)
	for {
		time.Sleep(aggregateInterval * time.Second)

		err := pa.syncAgentList()
		if err != nil {
			fmt.Println(err)
			return err
		}

		if len(pa.AgentList) == 0 {
			time.Sleep(aggregateInterval * time.Second)
			continue
		}

		go pa.AggregateMetric(pa.AgentList, metricArr, config.GetInstance().Monitoring.AggregateType)

	}
}

func (pa *PullAggregator) AggregateMetric(agentList map[string]metadata.AgentInfo, metricArr []types.Metric, aggregateType string) {
	for _, agent := range agentList {
		for _, metricKind := range metricArr {
			//receivedMetric, err := pa.Storage.ReadMetric(config.GetInstance().Monitoring.DefaultPolicy == types.PUSH_POLICY, agent.AgentID, metricKind.ToString(), "m", aggregateType, "5m")
			var receivedMetric interface{}
			var err error
			//var calculatedMetric interface{}
			var mappedMetric interface{}
			receivedMetric, err = pa.Storage.ReadMetric(config.GetInstance().Monitoring.DefaultPolicy == types.PUSH_POLICY, agent.VmId, metricKind.ToAgentMetricKey(), "m", aggregateType, "5m")
			if err != nil {
				logrus.Println(err)
			}
			if receivedMetric == nil {
				continue
			}
			mappedMetric, err = influxdb.MappingMonMetric(metricKind.ToString(), &receivedMetric)
			var metricName string
			var valueLength float64
			tagArr := map[string]string{}
			reqValue := map[string]interface{}{}
			if metricKind.ToString() == types.Network.ToString() || metricKind.ToString() == types.DiskIO.ToString() {
				for k, v := range mappedMetric.(map[string]interface{}) {
					if k == "values" {
						for _, vv := range v.([]interface{}) {
							for metricKey, metricValue := range vv.(map[string]interface{}) {
								valueLength += 1
								if metricKey == "time" {
									continue
								}
								compare, _ := metricValue.(json.Number).Float64()
								if reqValue[metricKey] == nil || aggregateType == types.LAST.ToString() {
									reqValue[metricKey] = compare
								} else {
									origin, _ := reqValue[metricKey].(float64)
									var vSum float64
									if aggregateType == types.MAX.ToString() {
										if origin < compare {
											reqValue[metricKey] = compare
										}
									}
									if aggregateType == types.MIN.ToString() {
										if origin > compare {
											reqValue[metricKey] = compare
										}
									}
									if aggregateType == types.AVG.ToString() {
										vSum += compare
										reqValue[metricKey] = vSum / valueLength

									}
								}
							}
						}
					}
					if k == "name" {
						metricName = v.(string)
					}
					if k == "tags" {
						for tKey, tValue := range v.(map[string]string) {
							tagArr[tKey] = tValue
						}
					}
				}
			} else {
				convertedMetric := mappedMetric.(map[string]interface{})
				metricName = convertedMetric["name"].(string)
				for k, v := range convertedMetric["tags"].(map[string]string) {
					tagArr[k] = v
				}
				for _, value := range convertedMetric["values"].([]interface{}) {
					for k, v := range value.(map[string]interface{}) {
						if k == "time" {
							continue
						}
						if v == nil {
							v = json.Number("0")
						}
						inputData, _ := v.(json.Number).Float64()
						reqValue[k] = inputData

					}
				}

			}
			err = influxdbv1.GetInstance().WriteOnDemandMetric(influxdbv1.DefaultDatabase, metricName, tagArr, reqValue)
			if err != nil {
				logrus.Println(err)
			}
			// TODO Aggregate 이후 데이터베이스의 이전 메트릭 데이터 삭제 로직 추가 필요
		}
	}
}

func (pa *PullAggregator) CalculateMetric() (map[string]interface{}, error) {
	return nil, nil
}

func (pa *PullAggregator) syncAgentList() error {
	syncedAgentList, err := pa.AgentListManager.GetAgentList()
	if err != nil {
		fmt.Println(err)
		return err
	}
	pa.AgentList = syncedAgentList
	return nil
}
