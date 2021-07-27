package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"sort"
	"strconv"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/cbstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdb/v1"
	"github.com/sirupsen/logrus"
	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/cloud-barista/cb-dragonfly/pkg/types"
)

type Aggregator struct {
	AggregateType types.AggregateType
}

func (a *Aggregator) AggregateMetric(kafkaConn *kafka.Consumer, topics []string) ([]string, error) {

	currentTime := time.Now().Unix()
	stayConnCount := 0

	var msgSlice [][]byte
	var msgTopic []string
	for {
		stayConnCount += 1
		msg, err := kafkaConn.ReadMessage(1 * time.Second)
		if err != nil {
			logrus.Debug("From AggregateMetric, pre-topics conn based kafkaConn bring about above err : ", err)
		}
		if msg != nil {
			msgTime := msg.Timestamp.Unix()
			msgSlice = append(msgSlice, msg.Value)
			msgTopic = append(msgTopic, *msg.TopicPartition.Topic)
			if msgTime > currentTime {
				break
			}
			stayConnCount = 0
			msg = nil
		}
		if stayConnCount == types.ReadConnectionTimeout {
			break
		}
	}
	fmt.Println(fmt.Sprintf("%v : %d\n", topics, len(msgSlice)))

	tagInfo := make(map[string]map[string]string)
	if len(msgSlice) != 0 {
		uniqueResponseSlice := make(map[string]map[string]map[string][]float64)
		for idx, value := range msgSlice {
			response := TelegrafMetric{}
			_ = json.Unmarshal(value, &response)

			vmTopic := msgTopic[idx]
			if _, ok := tagInfo[vmTopic]; ok {
				for key, tag := range response.Tags {
					if key == types.NsId || key == types.McisId || key == types.VmId || key == types.OsType || key == types.CspType {
						tagInfo[vmTopic][key] = tag.(string)
					}
				}
			} else {
				tagInfo[vmTopic] = make(map[string]string)
				for key, tag := range response.Tags {
					if key == types.NsId || key == types.McisId || key == types.VmId || key == types.OsType || key == types.CspType {
						tagInfo[vmTopic][key] = tag.(string)
					}
				}
			}

			if _, ok := uniqueResponseSlice[vmTopic]; ok {
				if _, ok := uniqueResponseSlice[vmTopic][response.Name]; ok {
					for metricName, val := range response.Fields {
						uniqueResponseSlice[vmTopic][response.Name][metricName] = append(uniqueResponseSlice[vmTopic][response.Name][metricName], val.(float64))
					}
				} else {
					uniqueResponseSlice[vmTopic][response.Name] = make(map[string][]float64)
					for metricName, val := range response.Fields {
						uniqueResponseSlice[vmTopic][response.Name][metricName] = append(uniqueResponseSlice[vmTopic][response.Name][metricName], val.(float64))
					}
				}
			} else {
				uniqueResponseSlice[vmTopic] = make(map[string]map[string][]float64)
				uniqueResponseSlice[vmTopic][response.Name] = make(map[string][]float64)
				for metricName, val := range response.Fields {
					uniqueResponseSlice[vmTopic][response.Name][metricName] = append(uniqueResponseSlice[vmTopic][response.Name][metricName], val.(float64))
				}
			}
		}
		result, err := a.CalculateMetric(uniqueResponseSlice, tagInfo, a.AggregateType.ToString())
		if err != nil {
			util.GetLogger().Error(err)
		}
		err = v1.GetInstance().WriteMetric(v1.DefaultDatabase, result)
		if err != nil {
			return []string{}, err
		}
	}

	currentTopics := unique(msgTopic)
	for _, topic := range currentTopics {
		cbstore.GetInstance().StoreDelete(types.DELTOPICS + topic)
	}
	delTopics := []string{}
	needCheckTopics := ReturnDiffTopicList(topics, currentTopics)
	if len(needCheckTopics) != 0 {
		for _, topic := range needCheckTopics {
			if cbstore.GetInstance().StoreGet(types.DELTOPICS+topic) == "" {
				cbstore.GetInstance().StorePut(types.DELTOPICS+topic, "0")
			} else {
				count, _ := strconv.Atoi(cbstore.GetInstance().StoreGet(types.DELTOPICS + topic))
				count++
				cbstore.GetInstance().StorePut(types.DELTOPICS+topic, strconv.Itoa(count))
			}
			checkNum, _ := strconv.Atoi(cbstore.GetInstance().StoreGet(types.DELTOPICS + topic))
			if checkNum >= 2 {
				delTopics = append(delTopics, topic)
			}
		}
	}

	return ReturnDiffTopicList(topics, delTopics), nil
}

func (a *Aggregator) CalculateMetric(responseMap map[string]map[string]map[string][]float64, tagMap map[string]map[string]string, aggregateType string) (map[string]interface{}, error) {

	resultMap := map[string]interface{}{}

	for vmTopic, metric := range responseMap {
		resultMap[vmTopic] = make(map[string]interface{})
		for metricName, metricSlice := range metric {
			metric := map[string]interface{}{}
			for key, slice := range metricSlice {
				switch types.AggregateType(aggregateType) {
				case types.MIN:
					sort.Sort(sort.Float64Slice(slice))
					metric[key] = slice[0]
				case types.MAX:
					sort.Sort(sort.Reverse(sort.Float64Slice(slice)))
					metric[key] = slice[0]
				case types.AVG:
					var sum float64
					for _, v := range slice {
						sum += v
					}
					metric[key] = sum / float64(len(slice))
				case types.LAST:
					metric[key] = slice[len(slice)-1]
				}
				resultMap[vmTopic].(map[string]interface{})[metricName] = metric
			}
			resultMap[vmTopic].(map[string]interface{})["tagInfo"] = tagMap[vmTopic]
		}
	}
	for topic, topicVal := range resultMap {
		for metricName, metricVal := range topicVal.(map[string]interface{}) {
			if metricName == "tagInfo" {
				continue
			}
			tmpMetric, err := mappingOnDemandMetric(config.GetInstance().Monitoring.DefaultPolicy == types.PushPolicy, types.GetMetricType(metricName), metricVal.(map[string]interface{}))
			if err != nil {
				util.GetLogger().Error(err)
				return nil, err
			}
			resultMap[topic].(map[string]interface{})[metricName] = tmpMetric
		}
	}
	return resultMap, nil
}

func ConvertMonMetric(metric types.Metric, metricVal TelegrafMetric) (map[string]interface{}, error) {
	metricMap := map[string]interface{}{}
	metricMap["name"] = metricVal.Name
	tagMap := map[string]interface{}{
		"nsId":    metricVal.Tags["nsId"],
		"mcisId":  metricVal.Tags["mcisId"],
		"vmId":    metricVal.Tags["vmId"],
		"osType":  metricVal.Tags["osType"],
		"cspType": metricVal.Tags["cspType"],
	}
	metricMap["tags"] = tagMap

	metricCols, err := mappingOnDemandMetric(config.GetInstance().Monitoring.DefaultPolicy == types.PushPolicy, metric, metricVal.Fields)
	if err != nil {
		return nil, err
	}
	metricMap["values"] = metricCols
	metricMap["time"] = time.Now().UTC()
	return metricMap, nil
}

func mappingOnDemandMetric(isPush bool, metric types.Metric, metricVal map[string]interface{}) (map[string]interface{}, error) {
	// Metric 구조체 Map 변환
	metricCols := map[string]interface{}{}

	switch metric {
	case types.Cpu:
		metricCols["cpu_utilization"] = metricVal["usage_utilization"]
		metricCols["cpu_system"] = metricVal["usage_system"]
		metricCols["cpu_idle"] = metricVal["usage_idle"]
		metricCols["cpu_iowait"] = metricVal["usage_iowait"]
		metricCols["cpu_hintr"] = metricVal["usage_irq"]
		metricCols["cpu_sintr"] = metricVal["usage_softirq"]
		metricCols["cpu_user"] = metricVal["usage_user"]
		metricCols["cpu_nice"] = metricVal["usage_nice"]
		metricCols["cpu_steal"] = metricVal["usage_steal"]
		metricCols["cpu_guest"] = metricVal["usage_guest"]
		metricCols["cpu_guest_nice"] = metricVal["usage_guest_nice"]
	case types.CpuFrequency:
		metricCols["cpu_speed"] = metricVal["cur_freq"]
	case types.Memory:
		metricCols["mem_utilization"] = metricVal["used_percent"]
		metricCols["mem_total"] = metricVal["total"]
		metricCols["mem_used"] = metricVal["used"]
		metricCols["mem_free"] = metricVal["free"]
		metricCols["mem_shared"] = metricVal["shared"]
		metricCols["mem_buffers"] = metricVal["buffered"]
		metricCols["mem_cached"] = metricVal["cached"]
	case types.Disk:
		metricCols["disk_utilization"] = metricVal["used_percent"]
		metricCols["disk_total"] = metricVal["total"]
		metricCols["disk_used"] = metricVal["used"]
		metricCols["disk_free"] = metricVal["free"]
	case types.DiskIO:
		metricCols["kb_read"] = metricVal["read_bytes"]
		metricCols["kb_written"] = metricVal["write_bytes"]
		metricCols["ops_read"] = metricVal["iops_read"]
		metricCols["ops_write"] = metricVal["iops_write"]
	case types.Network:
		metricCols["bytes_in"] = metricVal["bytes_recv"]
		metricCols["bytes_out"] = metricVal["bytes_sent"]
		metricCols["pkts_in"] = metricVal["packets_recv"]
		metricCols["pkts_out"] = metricVal["packets_sent"]
	case types.None:
	default:
		if isPush {
			break
		}
		err := errors.New("not found metric")
		return nil, err
	}

	return metricCols, nil
}
