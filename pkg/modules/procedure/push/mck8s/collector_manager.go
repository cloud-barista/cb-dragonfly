package mcis

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/modules/procedure/push/mck8s/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
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

	deployType := config.GetInstance().Monitoring.DeployType
	if deployType == types.Dev || deployType == types.Compose {
		go func() {
			err := newCollector.DoCollect(manager.WaitGroup)
			if err != nil {
				errMsg := fmt.Sprintf("failed to create collector, error=%s", err.Error())
				util.GetLogger().Error(errMsg)
			}
		}()
	}

	return nil
}

// DeleteCollector 콜렉터 삭제
func (manager *CollectManager) DeleteCollector(topic string) error {
	if _, ok := manager.CollectorAddrMap[topic]; !ok {
		return errors.New(fmt.Sprint("failed to find collector with topic", topic))
	}

	targetCollector := manager.CollectorAddrMap[topic]
	deployType := config.GetInstance().Monitoring.DeployType
	if deployType == types.Dev || deployType == types.Compose {
		// 콜렉터 채널에 종료 요청
		targetCollector.Ch <- "close"
	}

	// 콜렉터 목록에서 콜렉터 삭제
	delete(manager.CollectorAddrMap, topic)

	return nil
}
