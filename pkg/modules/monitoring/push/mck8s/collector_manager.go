package mcis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/modules/monitoring/push/mck8s/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CollectManager struct {
	CollectorAddrMap map[string]*collector.MetricCollector
	CollectorPolicy  string
	K8sClientSet     *kubernetes.Clientset
	WaitGroup        *sync.WaitGroup
}

func NewCollectorManager(wg *sync.WaitGroup) (*CollectManager, error) {
	manager := CollectManager{}
	if config.GetInstance().Monitoring.DeployType == types.Helm {
		if err := manager.InitDFK8sEnv(); err != nil {
			return &manager, err
		}
	}
	//CHECK: Helm 일 경우 아래 로직은 왜 안타도 되는게 맞는지 확인 필요
	manager.CollectorAddrMap = map[string]*collector.MetricCollector{}
	manager.CollectorPolicy = strings.ToUpper(config.GetInstance().Monitoring.MonitoringPolicy)
	manager.WaitGroup = wg
	return &manager, nil
}

func (manager *CollectManager) InitDFK8sEnv() (err error) {
	// k8s conn set Start
	inClusterK8sConfig, err := rest.InClusterConfig()
	clientSet, err := kubernetes.NewForConfig(inClusterK8sConfig)
	manager.K8sClientSet = clientSet
	// k8s conn set End

	// helm 으로 배포할 경우, df 는 collector 를 deployment 형태로 배포합니다.
	// df 와 collector 는 configmap 으로 topic 정보를 동기화합니다.
	// 아래 코드는 configmap 을 설정 및 배포하는 코드입니다.
	configMapsClient := manager.K8sClientSet.CoreV1().ConfigMaps(config.GetInstance().Dragonfly.HelmNamespace)
	configMap := &apiv1.ConfigMap{Data: map[string]string{}, ObjectMeta: metav1.ObjectMeta{
		Name: types.MCK8SConfigMapName,
	}}
	// Deploy ConfigMap => (1) 드래곤 플라이가 배포한 컨피그맵이 이미 생성되어 있는지 조회 (2) 컨피그 맵이 없을 경우, 배포
	_, notExistErr := configMapsClient.Get(context.TODO(), types.MCK8SConfigMapName, metav1.GetOptions{})
	if notExistErr != nil {
		fmt.Println("configmap creating")
		result, err := configMapsClient.Create(context.TODO(), configMap, metav1.CreateOptions{})
		if err != nil {
			fmt.Println("configmap create error: ", err)
			return err
		}
		fmt.Println("Created ConfigMap: ", result.GetObjectMeta().GetName())
	}
	return
}

// CreateCollector 콜렉터 생성
func (manager *CollectManager) CreateCollector(topic string) error {
	manager.WaitGroup.Add(1)
	collectorCreateOrder := len(manager.CollectorAddrMap)
	newCollector, err := collector.NewMetricCollector(
		types.AggregateType(config.GetInstance().Monitoring.AggregateType),
		collectorCreateOrder,
	)
	if err != nil {
		return err
	}

	manager.CollectorAddrMap[topic] = &newCollector

	switch config.GetInstance().Monitoring.DeployType {
	case types.Helm:
		collectorUUID := fmt.Sprintf("%p", &newCollector)
		env := []apiv1.EnvVar{
			{Name: "topic", Value: topic},
			{Name: "kafka_endpoint_url", Value: config.GetInstance().Kafka.EndpointUrl},
			{Name: "create_order", Value: strconv.Itoa(collectorCreateOrder)},
			{Name: "namespace", Value: config.GetInstance().Dragonfly.HelmNamespace},
			{Name: "df_addr", Value: fmt.Sprintf("%s:%d", config.GetInstance().Dragonfly.DragonflyIP, config.GetInstance().Dragonfly.HelmPort)},
			{Name: "mck8s_collector_interval", Value: strconv.Itoa(config.GetInstance().Monitoring.MCK8SCollectorInterval)},
			{Name: "collect_uuid", Value: collectorUUID},
		}
		deploymentTemplate := util.DeploymentTemplate(types.MCK8SDeploymentName, collectorCreateOrder, collectorUUID, env, types.MCK8SCollectorImage)
		fmt.Println("Creating deployment...")
		result, err := manager.K8sClientSet.AppsV1().Deployments(config.GetInstance().Dragonfly.HelmNamespace).Create(context.TODO(), deploymentTemplate, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		fmt.Println("Created deployment: ", result.GetObjectMeta().GetName())
		return nil
	case types.Dev, types.Compose:
		go func() {
			err := newCollector.DoCollect(manager.WaitGroup)
			if err != nil {
				errMsg := fmt.Sprintf("failed to create collector, error=%s", err.Error())
				util.GetLogger().Error(errMsg)
			}
		}()
	}

	defer func(topicData string) {
		curTime := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] <MCK8S> Create collector - topic: %s\n", curTime, topicData)
	}(topic)

	return nil
}

// DeleteCollector 콜렉터 삭제
func (manager *CollectManager) DeleteCollector(topic string) error {
	if _, ok := manager.CollectorAddrMap[topic]; !ok {
		return errors.New(fmt.Sprint("failed to find collector with topic", topic))
	}

	targetCollector := manager.CollectorAddrMap[topic]
	switch config.GetInstance().Monitoring.DeployType {
	case types.Dev, types.Compose:
		// 콜렉터 채널에 종료 요청
		targetCollector.Ch <- "close"
	}

	defer func(topicData string) {
		curTime := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] <MCK8S> Delete collector - topic: %s\n", curTime, topicData)
	}(topic)
	// 콜렉터 목록에서 콜렉터 삭제
	delete(manager.CollectorAddrMap, topic)

	return nil
}
