package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type MetricCollector struct {
	CreateOrder       int
	ConsumerKafkaConn *kafka.Consumer
	Aggregator        Aggregator
	Ch                chan []string
}

var KafkaConfig *kafka.ConfigMap

func NewMetricCollector(aggregateType types.AggregateType, createOrder int) (MetricCollector, error) {
	//KafkaConfig = &kafka.ConfigMap{
	//	"bootstrap.servers":  fmt.Sprintf("%s", config.GetDefaultConfig().Kafka.EndpointUrl),
	//	"group.id":           fmt.Sprintf("%d", createOrder),
	//	"enable.auto.commit": true,
	//	"auto.offset.reset":  "earliest",
	//}
	//
	//consumerKafkaConn, err := kafka.NewConsumer(KafkaConfig)
	//if err != nil {
	//	util.GetLogger().Error("Fail to create collector kafka consumer, Kafka Connection Fail", err)
	//	util.GetLogger().Error(err)
	//	return MetricCollector{}, err
	//}
	//ch := make(chan []string)
	mc := MetricCollector{
		//ConsumerKafkaConn: consumerKafkaConn,
		CreateOrder: createOrder,
		//Aggregator: Aggregator{
		//	AggregateType: aggregateType,
		//},
		//Ch: ch,
	}
	fmt.Println(fmt.Sprintf("#### Group_%d collector Create ####", createOrder))
	return mc, nil
}

func (mc *MetricCollector) DoCollect(wg *sync.WaitGroup) error {
	defer wg.Done()
	for {
		time.Sleep(5 * time.Second)
		fmt.Printf("[%d] DoCollect ...\n", mc.CreateOrder)
	}
	return nil
}
