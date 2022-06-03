package collector

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type MetricCollector struct {
	CreateOrder       int
	ConsumerKafkaConn *kafka.Consumer
	Aggregator        Aggregator
	Ch                chan string
}

func (mc *MetricCollector) DoCollect(wg *sync.WaitGroup) error {
	defer wg.Done()
	for {
		select {
		case chanData := <-mc.Ch:
			if len(chanData) != 0 {

				// 콜렉터 삭제 요청일 경우 토픽 구독 취소 및 삭제 처리
				if chanData == "close" {
					close(mc.Ch)
					_ = mc.ConsumerKafkaConn.Unsubscribe()
					err := mc.ConsumerKafkaConn.Close()
					if err != nil {
						errMsg := fmt.Sprintf("fail to delete mcks metic collector, kafka close connection failed with error=%s", err)
						util.GetLogger().Error(errMsg)
						fmt.Println(errMsg)
						return errors.New(errMsg)
					}
					fmt.Println(fmt.Sprintf("#### Group_%d MCKS collector Delete ####", mc.CreateOrder))
					return nil
				}

				deliveredTopic := chanData
				fmt.Println(fmt.Sprintf("Group_%d MCKS collector Delivered : %s", mc.CreateOrder, deliveredTopic))

				// 토픽 데이터 구독
				err := mc.ConsumerKafkaConn.SubscribeTopics([]string{deliveredTopic}, nil)
				if err != nil {
					errMsg := fmt.Sprintf("fail to subscribe topic with topic %s, error=%s", deliveredTopic, err)
					util.GetLogger().Error(errMsg)
					return errors.New(errMsg)
				}

				// TODO: 토픽 데이터 처리
				mc.Aggregator.AggregateMetric(mc.ConsumerKafkaConn, deliveredTopic)

				// TODO: 비활성화된 토픽 체크

			}
			break
		}
	}
}

func NewMetricCollector(aggregateType types.AggregateType, createOrder int) (MetricCollector, error) {

	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":  fmt.Sprintf("%s", config.GetDefaultConfig().Kafka.EndpointUrl),
		"group.id":           fmt.Sprintf("%d", createOrder),
		"enable.auto.commit": true,
		"auto.offset.reset":  "earliest",
	}
	consumerKafkaConn, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		errMsg := fmt.Sprintf("fail to create mcks metic collector, kafka connection failed with error=%s", err)
		util.GetLogger().Error(errMsg)
		return MetricCollector{}, errors.New(errMsg)
	}

	ch := make(chan string)
	mc := MetricCollector{
		ConsumerKafkaConn: consumerKafkaConn,
		CreateOrder:       createOrder,
		Aggregator: Aggregator{
			AggregateType: aggregateType,
		},
		Ch: ch,
	}
	fmt.Println(fmt.Sprintf("#### Group_%d MCKS collector Create ####", createOrder))
	return mc, nil
}
