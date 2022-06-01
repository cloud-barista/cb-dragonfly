package mcis

import (
	"strings"
	"sync"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/modules/procedure/push/mcks/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
)

type CollectManager struct {
	CollectorAddrMap map[string]*collector.MetricCollector
	CollectorPolicy  string
	WaitGroup        *sync.WaitGroup
}

func NewCollectorManager(wg *sync.WaitGroup) (CollectManager, error) {
	manager := CollectManager{
		CollectorAddrMap: map[string]*collector.MetricCollector{},
		CollectorPolicy:  strings.ToUpper(config.GetInstance().Monitoring.MonitoringPolicy),
		WaitGroup:        wg,
	}
	return manager, nil
}

func (manager *CollectManager) CreateCollector(topic string) error {
	// collector 생성 요청이 들어왔을 경우
	manager.WaitGroup.Add(1)
	collectorCreateOrder := len(manager.CollectorAddrMap)
	// 생성 순서 idx 값을 collector 객체에 넣고, 생성합니다.
	newCollector, err := collector.NewMetricCollector(
		types.AVG,
		collectorCreateOrder,
	)
	if err != nil {
		return err
	}
	// 생성한 DoCollect 의 주소 값을 CollectorAddrSlice 배열에 추가합니다.
	manager.CollectorAddrMap[topic] = &newCollector

	deployType := config.GetInstance().Monitoring.DeployType
	if deployType == types.Dev || deployType == types.Compose {
		go func() {
			err := newCollector.DoCollect(manager.WaitGroup)
			if err != nil {
				//util.GetLogger().Error("failed to DoCollect")
			}
		}()
	}

	return nil
}

func (manager *CollectManager) DeleteCollector(topic string) error {
	return nil
}
