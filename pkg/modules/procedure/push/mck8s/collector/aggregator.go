package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v1 "github.com/cloud-barista/cb-dragonfly/pkg/storage/metricstore/influxdb/v1"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/thoas/go-funk"
)

type TelegrafMetric struct {
	Name      string                 `json:"name"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
	Timestamp int64                  `json:"timestamp"`
}

type Aggregator struct {
	AggregateType types.AggregateType
}

func (a *Aggregator) AggregateMetric(kafkaConn *kafka.Consumer, topic string) (bool, error) {
	curTime := time.Now().Unix()
	reconnectTry := 0
	var topicMsgBytes [][]byte

	// 토픽 메세지 조회
	for {
		reconnectTry++

		topicMsg, err := kafkaConn.ReadMessage(5 * time.Second)
		if err != nil {
			errMsg := fmt.Sprintf("fail to read topic message with topic %s, error=%s", topic, err)
			util.GetLogger().Error(errMsg)
			continue
		}
		if topicMsg != nil {
			// 토픽 메세지 저장
			topicMsgBytes = append(topicMsgBytes, topicMsg.Value)
			// 토픽 생성 시간 체크
			if topicMsg.Timestamp.Unix() > curTime {
				break
			}
			// 토픽 메세지 및 타임아웃 정보 초기화
			reconnectTry = 0
			topicMsg = nil
		}

		// 토픽 타임아웃 재시도 횟수 제한 체크
		if reconnectTry >= types.ReadConnectionTimeout {
			errMsg := fmt.Sprintf("exceed maximum kafka connection %d", reconnectTry)
			util.GetLogger().Error(errMsg)
			break
		}
	}

	if len(topicMsgBytes) == 0 {
		return false, errors.New("failed to get monitoring data from kafka, data bytes is zero")
	}

	// TODO: 최초 토픽 데이터 처리 시 에이전트 메타데이터 헬스상태 변경

	// 토픽 메세지 파싱
	metrics := make([]TelegrafMetric, len(topicMsgBytes))
	for idx, topicMsgByte := range topicMsgBytes {
		var metric TelegrafMetric
		err := json.Unmarshal(topicMsgByte, &metric)
		if err == nil {
			metrics[idx] = metric
		}
	}

	// 모니터링 데이터 처리
	a.Aggregate(metrics)

	return true, nil
}

// Aggregate 쿠버네티스 노드, 파드 메트릭 처리 및 저장
func (a *Aggregator) Aggregate(metrics []TelegrafMetric) {

	// 쿠버네티스 노드 메트릭 처리 및 저장
	a.aggregateNodeMetric(metrics)

	// 쿠버네티스 파드 메트릭 처리 및 저장
	a.aggregatePodMetric(metrics, "kubernetes_pod_container")

	// 쿠버네티스 파드 네트워크 메트릭 처리 및 저장
	a.aggregatePodMetric(metrics, "kubernetes_pod_network")
}

// aggregateNodeMetric 쿠버네티스 노드 메트릭 처리 및 저장
func (a *Aggregator) aggregateNodeMetric(metrics []TelegrafMetric) {
	metricName := "kubernetes_node"

	// 1. 토픽 메세지에서 노드 메트릭 메세지를 필터링
	nodeMetricFilter := funk.Filter(metrics, func(metric TelegrafMetric) bool {
		return metric.Name == "kubernetes_node"
	})
	nodeMetricArr := nodeMetricFilter.([]TelegrafMetric)
	if len(nodeMetricArr) != 0 {
		// 2. 전체 클러스터에서 노드 목록 추출
		nodeNameFilter := funk.Uniq(funk.Get(nodeMetricArr, "Tags.node_name"))
		nodeNameArr := nodeNameFilter.([]string)

		// 개별 노드에 대한 모니터링 메트릭 처리 및 저장
		for _, nodeName := range nodeNameArr {
			currentNodeMetricArr := funk.Filter(nodeMetricArr, func(metric TelegrafMetric) bool {
				return metric.Tags["node_name"] == nodeName
			})
			nodeMetric := aggregateMetric(metricName, currentNodeMetricArr.([]TelegrafMetric), string(a.AggregateType))
			err := v1.GetInstance().WriteOnDemandMetric(v1.DefaultDatabase, nodeMetric.Name, nodeMetric.Tags, nodeMetric.Fields)
			if err != nil {
				util.GetLogger().Error(fmt.Sprintf("failed to write metric, error=%s", err.Error()))
				continue
			}
		}
	}
}

// aggregateMetric 쿠버네티스 파드 메트릭 처리 및 저장
func (a *Aggregator) aggregatePodMetric(metrics []TelegrafMetric, metricName string) {
	// 1. 토픽 메세지에서 파드 메트릭 메세지를 필터링
	podMetricFilter := funk.Filter(metrics, func(metric TelegrafMetric) bool {
		return metric.Name == metricName
	})
	podMetricArr := podMetricFilter.([]TelegrafMetric)

	// 2. 파드 별 메트릭 처리
	if podMetricArr != nil {
		podNameFilter := funk.Uniq(funk.Get(podMetricArr, "Tags.pod_name"))
		podNameArr := podNameFilter.([]string)

		// 3. 단일 파드에 대한 파드 메트릭 처리 및 저장
		for _, podName := range podNameArr {
			currentPodMetricArr := funk.Filter(podMetricArr, func(metric TelegrafMetric) bool {
				return metric.Tags["pod_name"] == podName
			})
			podMetric := aggregateMetric(metricName, currentPodMetricArr.([]TelegrafMetric), string(a.AggregateType))
			err := v1.GetInstance().WriteOnDemandMetric(v1.DefaultDatabase, podMetric.Name, podMetric.Tags, podMetric.Fields)
			if err != nil {
				util.GetLogger().Error(fmt.Sprintf("failed to write metric, error=%s", err.Error()))
				continue
			}
		}
	}
}

// aggregateMetric 쿠버네티스 메트릭 처리 및 저장
func aggregateMetric(metricType string, metrics []TelegrafMetric, criteria string) TelegrafMetric {
	aggregatedMetric := TelegrafMetric{
		Name:   metricType,
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}

	// 모니터링 메트릭 태그 정보 설정
	metricTags := make(map[string]string)
	for k, v := range metrics[0].Tags {
		metricTags[k] = v
	}
	aggregatedMetric.Tags = metricTags

	// last
	if criteria == "last" {
		// 가장 최신의 타임스탬프 값 가져오기
		timestampArr := funk.Get(metrics, "Timestamp")
		latestTimestamp := funk.MaxInt64(timestampArr.([]int64))
		// 최신 타임스탬프 값의 모니터링 메트릭 조회
		latestMetric := funk.Filter(metrics, func(metric TelegrafMetric) bool {
			return metric.Timestamp == latestTimestamp
		})
		return latestMetric.(TelegrafMetric)
	}

	// min, max, avg
	fieldList := funk.Keys(metrics[0].Fields)
	for _, fieldKey := range fieldList.([]string) {
		// 필드 별 데이터만 추출
		fieldDataArr := funk.Get(metrics, fmt.Sprintf("Fields.%s", fieldKey))
		// min, max, avg, last 처리
		var aggregatedVal float64
		if criteria == "min" {
			aggregatedVal = funk.MinFloat64(convertToFloat64Arr(fieldDataArr))
		} else if criteria == "max" {
			aggregatedVal = funk.MaxFloat64(convertToFloat64Arr(fieldDataArr))
		} else if criteria == "avg" {
			aggregatedVal = funk.SumFloat64(convertToFloat64Arr(fieldDataArr))
			aggregatedVal = aggregatedVal / float64(len(fieldDataArr.([]interface{})))
		}
		aggregatedMetric.Fields[fieldKey] = aggregatedVal
	}
	return aggregatedMetric
}

func convertToFloat64Arr(interfaceArr interface{}) []float64 {
	var float64Arr []float64
	for _, elem := range interfaceArr.([]interface{}) {
		float64Arr = append(float64Arr, elem.(float64))
	}
	return float64Arr
}
