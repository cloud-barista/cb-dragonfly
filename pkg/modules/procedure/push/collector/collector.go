package collector

import (
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/sirupsen/logrus"
	"sort"
	"sync"

	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/go-cmp/cmp"
)

type MetricCollector struct {
	CreateOrder       int
	ConsumerKafkaConn *kafka.Consumer
	Aggregator        Aggregator
	Ch                chan []string
}

var KafkaConfig *kafka.ConfigMap

func NewMetricCollector(aggregateType types.AggregateType, createOrder int) (MetricCollector, error) {

	KafkaConfig = &kafka.ConfigMap{
		"bootstrap.servers":  fmt.Sprintf("%s", config.GetDefaultConfig().Kafka.EndpointUrl),
		"group.id":           fmt.Sprintf("%d", createOrder),
		"enable.auto.commit": true,
		"session.timeout.ms": 15000,
		"auto.offset.reset":  "earliest",
	}

	consumerKafkaConn, err := kafka.NewConsumer(KafkaConfig)
	if err != nil {
		util.GetLogger().Error("Fail to create collector kafka consumer, Kafka Connection Fail", err)
		util.GetLogger().Error(err)
		return MetricCollector{}, err
	}
	ch := make(chan []string)
	mc := MetricCollector{
		ConsumerKafkaConn: consumerKafkaConn,
		CreateOrder:       createOrder,
		Aggregator: Aggregator{
			AggregateType: aggregateType,
		},
		Ch: ch,
	}
	fmt.Println(fmt.Sprintf("#### Group_%d collector Create ####", createOrder))
	return mc, nil
}

func (mc *MetricCollector) Collector(wg *sync.WaitGroup) error {

	defer wg.Done()
	aliveTopics := []string{}
	for {
		select {
		case processDecision := <-mc.Ch:
			if len(processDecision) != 0 {

				if processDecision[0] == "close" {
					close(mc.Ch)
					_ = mc.ConsumerKafkaConn.Unsubscribe()
					err := mc.ConsumerKafkaConn.Close()
					if err != nil {
						logrus.Debug("Fail to collector kafka connection close")
					}
					fmt.Println(fmt.Sprintf("#### Group_%d collector Delete ####", mc.CreateOrder))
					return nil
				}

				DeliveredTopicList := processDecision
				fmt.Println(fmt.Sprintf("Group_%d collector Delivered : %s", mc.CreateOrder, DeliveredTopicList))

				sort.Strings(aliveTopics)
				topicList := util.ReturnDiffTopicList(DeliveredTopicList, aliveTopics)
				if len(topicList) != 0 {
					err := mc.ConsumerKafkaConn.SubscribeTopics(DeliveredTopicList, nil)
					if err != nil {
						fmt.Println(err)
					}
				}

				aliveTopics, _ = mc.Aggregator.AggregateMetric(mc.ConsumerKafkaConn, DeliveredTopicList)
				if !cmp.Equal(DeliveredTopicList, aliveTopics) {
					_ = mc.ConsumerKafkaConn.Unsubscribe()
					deadTopics := util.ReturnDiffTopicList(DeliveredTopicList, aliveTopics)
					for _, delTopic := range deadTopics {
						if err := util.RingQueuePut(types.TopicDel, delTopic); err != nil {
							logrus.Debug(fmt.Sprintf("failed to put topics to ring queue, error=%s", err))
						}
					}
				}
			}
			break
		}
	}
}
