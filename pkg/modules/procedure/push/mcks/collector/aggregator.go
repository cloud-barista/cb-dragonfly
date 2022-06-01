package collector

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type TelegrafMetric struct {
	Name      string                 `json:"name"`
	Tags      map[string]interface{} `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
	Timestamp int64                  `json:"timestamp"`
	TagInfo   map[string]interface{} `json:"tagInfo"`
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
		topicMsg, err := kafkaConn.ReadMessage(5 * time.Second)
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

	// 토픽 메세지 처리

	return true, nil
}

func (a *Aggregator) CalculateMetric(responseMap map[string]map[string]map[string][]float64, tagMap map[string]map[string]string, aggregateType string) (map[string]interface{}, error) {
	return nil, nil
}
