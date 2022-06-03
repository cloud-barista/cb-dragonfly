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
	"github.com/davecgh/go-spew/spew"
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
		topicMsg, err := kafkaConn.ReadMessage(60 * time.Second)
		if err != nil {
			errMsg := fmt.Sprintf("fail to read topic message with topic %s, error=%s", topic, err)
			util.GetLogger().Error(errMsg)
			return false, errors.New(errMsg)
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

		// 토픽 타임아웃 체크
		if reconnectTry >= types.ReadConnectionTimeout {
			break
		}
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

	// 모니터링 데이터 가공 및 저장
	err := a.Aggregate(metrics)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *Aggregator) Aggregate(metrics []TelegrafMetric) error {

	/* 쿠버네티스 노드 메트릭 처리 및 저장 */

	// 1. 토픽 메세지에서 노드 메트릭 메세지를 필터링
	metricFilter := funk.Filter(metrics, func(metric TelegrafMetric) bool {
		return metric.Name == "kubernetes_node"
	})
	metricArr := metricFilter.([]TelegrafMetric)

	// 2. 클러스터 메트릭 처리
	// 전체 노드(클러스터)에 대한 노드 메트릭 처리 및 저장
	clusterMetric := aggregateNodeMetric(metricArr, "", string(a.AggregateType))
	spew.Dump(clusterMetric)
	err := v1.GetInstance().WriteOnDemandMetric(v1.DefaultDatabase, clusterMetric.Name, clusterMetric.Tags, clusterMetric.Fields)
	if err != nil {
		util.GetLogger().Error(fmt.Sprintf("failed to write metric, error=%s", err.Error()))
		return err
	}

	// 3. 노드 별 메트릭 처리
	// 클러스터 내 노드 목록 조회
	nodeNameFilter := funk.Uniq(funk.Get(metricArr, "Tags.node_name"))
	nodeNameArr := nodeNameFilter.([]string)

	// 단일 노드에 대한 노드 메트릭 처리 및 저장
	for _, nodeName := range nodeNameArr {
		currentNodeMetricArr := funk.Filter(metricArr, func(metric TelegrafMetric) bool {
			return metric.Tags["node_name"] == nodeName
		})
		nodeMetric := aggregateNodeMetric(currentNodeMetricArr.([]TelegrafMetric), nodeName, string(a.AggregateType))
		err := v1.GetInstance().WriteOnDemandMetric(v1.DefaultDatabase, nodeMetric.Name, nodeMetric.Tags, nodeMetric.Fields)
		if err != nil {
			util.GetLogger().Error(fmt.Sprintf("failed to write metric, error=%s", err.Error()))
			return err
		}
	}

	// TODO:
	// 4. 파드 별 메트릭 처리

	return nil
}

// aggregateNodeMetric 쿠버네티스 노드 메트릭 처리 및 저장
func aggregateNodeMetric(metrics []TelegrafMetric, nodeName string, criteria string) TelegrafMetric {
	aggregatedMetric := TelegrafMetric{
		Name:   "kubernetes_node",
		Tags:   map[string]string{},
		Fields: map[string]interface{}{},
	}

	// 노드 모니터링 메트릭 태그 정보 설정
	clusterTags := make(map[string]string)
	for k, v := range metrics[0].Tags {
		clusterTags[k] = v
	}
	if nodeName == "" {
		// 클러스터 전체 모니터링 메트릭일 경우 호스트, 노드 태그 삭제 처리
		delete(clusterTags, "host")
		delete(clusterTags, "node_name")
	}
	aggregatedMetric.Tags = clusterTags

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
