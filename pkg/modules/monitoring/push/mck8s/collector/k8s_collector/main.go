package main

import (
	"context"
	"encoding/json"
	"fmt"
	collector2 "github.com/cloud-barista/cb-dragonfly/pkg/modules/monitoring/push/mck8s/collector"
	"os"
	"strconv"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type MetricCollector struct {
	KafkaAdminClient  *kafka.AdminClient
	KafkaConsumerConn *kafka.Consumer
	CreateOrder       int
	Aggregator        collector2.Aggregator
	Ch                chan string
}

var KafkaConfig *kafka.ConfigMap

func PrintPanicError(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func DeleteDeployment(clientSet *kubernetes.Clientset, createOrder int, collectorUUID string, namespace string) {
	fmt.Println("MCK8S Deleting deployment...")
	deploymentName := fmt.Sprintf("%s%d-%s", types.MCK8SDeploymentName, createOrder, collectorUUID)
	deploymentsClient := clientSet.AppsV1().Deployments(namespace)
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Println("Fail to delete deployment.")
		fmt.Println(err)
	}
}

// deployment 로 배포된 collector
// 일정 주기( aggreTime )를 가지고 configmap 을 조회
// configmap 의 데이터( topicMaps ) 파싱하여, 자신의 collector idx 값을 가진 topics 들을 구독
// 만약 자신의 collector idx 값이 없다면 스스로 deployment 삭제 요청
func main() {
	fmt.Println("MCK8S main.go start")
	/** Get Env Val Start */
	kafkaEndpointUrl := os.Getenv("kafka_endpoint_url")
	var createOrder int
	createOrderString := os.Getenv("topic")
	if createOrderString == "" {
		fmt.Println("Get Env Error")
		return
	}
	createOrder, _ = strconv.Atoi(createOrderString)
	topic := os.Getenv("topic")
	aggregateType := types.AVG
	namespace := os.Getenv("namespace")
	dfAddr := os.Getenv("df_addr")
	collectInterval, _ := strconv.Atoi(os.Getenv("mck8s_collector_interval"))
	collectorUUID := os.Getenv("collect_uuid")
	if kafkaEndpointUrl == "" || namespace == "" || dfAddr == "" {
		fmt.Println("Get Env Error")
		return
	}
	/** Get Env Val End */

	/** Set Kafka, ConfigMap Conn Start */
	KafkaConfig = &kafka.ConfigMap{
		"bootstrap.servers":  kafkaEndpointUrl,
		"group.id":           fmt.Sprintf("%d", createOrder),
		"enable.auto.commit": true,
		"auto.offset.reset":  "earliest",
	}
	consumerKafkaConn, err := kafka.NewConsumer(KafkaConfig)
	PrintPanicError(err)
	config, errK8s := rest.InClusterConfig()
	PrintPanicError(errK8s)
	clientSet, errK8s2 := kubernetes.NewForConfig(config)
	PrintPanicError(errK8s2)
	/** Set Kafka, ConfigMap Conn End */

	/** Operate Collector Start */
	mc := MetricCollector{
		KafkaConsumerConn: consumerKafkaConn,
		CreateOrder:       createOrder,
		Aggregator: collector2.Aggregator{
			//CHECK: CreateOrder 가 굳이 Aggregator 에 있을 필요가 있을까?
			CreateOrder:   createOrder,
			AggregateType: aggregateType,
		},
	}
	fmt.Println(fmt.Sprintf("#### Group_%d collector Create ####", createOrder))

	configMapFailCnt := 0
	for {
		time.Sleep(time.Duration(collectInterval) * time.Second)
		fmt.Println(fmt.Sprintf("#### Group_%d collector ####", createOrder))
		fmt.Println("Get ConfigMap")
		/** Get ConfigMap<Data: Collector UUID Map, BinaryData: Collector Topics> Start */
		configMap, err := clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), types.MCK8SConfigMapName, metav1.GetOptions{})
		if err != nil {
			if configMapFailCnt == 5 {
				DeleteDeployment(clientSet, createOrder, collectorUUID, namespace)
			}
			configMapFailCnt += 1
			fmt.Println("Fail to Get ConfigMap")
			fmt.Println(err)
			continue
		}
		/** Get ConfigMap<Data: Collector UUID Map, BinaryData: Collector Topics> End */

		/** Check My Collector UUID Start */
		_, alive := configMap.Data[collectorUUID]
		if !alive {
			DeleteDeployment(clientSet, createOrder, collectorUUID, namespace)
		}
		/** Check My Collector UUID End */

		/** Get My Allocated Topics Start */
		topicMap := map[string][]string{}
		if err = json.Unmarshal(configMap.BinaryData["topicMap"], &topicMap); err != nil {
			fmt.Println("Fail to unMarshal ConfigMap Object Data")
		}
		deliveredTopic := topicMap[topic]
		fmt.Printf("[%s] <MCK8S> EXECUTE Group_%d collector - topic: %s\n", time.Now().Format(time.RFC3339), mc.CreateOrder, deliveredTopic)

		// 토픽 데이터 구독
		err = mc.KafkaConsumerConn.SubscribeTopics(deliveredTopic, nil)
		if err != nil {
			errMsg := fmt.Sprintf("fail to subscribe topic with topic %s, error=%s", deliveredTopic, err)
			util.GetLogger().Error(errMsg)
		}

		// 토픽 데이터 처리
		mc.Aggregator.AggregateMetric(mc.KafkaAdminClient, mc.KafkaConsumerConn, topic)
		/** Processing Topics to TSDB & Transmit Dead Topics To DF End */
	}
	/** Operate Collector End */
}
