package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/sirupsen/logrus"

	"github.com/cloud-barista/cb-dragonfly/pkg/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore/etcd"
)

// TODO: implements
// TODO: 1. API Server
// TODO: 2. Scheduling Collector...
// TODO: 3. Configuring Policy...

type CollectManager struct {
	//InfluxDB          metricstore.Storage
	//Etcd              realtimestore.Storage
	Aggregator        collector.Aggregator
	WaitGroup         *sync.WaitGroup
	UdpCOnn           *net.UDPConn
	metricL           *sync.RWMutex
	CollectorIdx      []string
	CollectorUUIDAddr map[string]*collector.MetricCollector
	AggregatingChan   map[string]*chan string
	TransmitDataChan  map[string]*chan collector.TelegrafMetric
	AgentQueueTTL     map[string]time.Time
	AgentQueueColN    map[string]int
}

// 콜렉터 매니저 초기화
func NewCollectorManager() (*CollectManager, error) {
	manager := CollectManager{}

	influxConfig := influxdbv1.Config{
		ClientOptions: []influxdbv1.ClientOptions{
			{
				URL:      fmt.Sprintf("%s:%d", config.GetInstance().GetInfluxDBConfig().EndpointUrl, config.GetInstance().GetInfluxDBConfig().InternalPort),
				Username: config.GetInstance().GetInfluxDBConfig().UserName,
				Password: config.GetInstance().GetInfluxDBConfig().Password,
			},
		},
		Database: config.GetInstance().GetInfluxDBConfig().Database,
	}

	// InfluxDB 연결
	err := influxdbv1.Initialize(influxConfig)
	if err != nil {
		logrus.Error("Failed to initialize influxDB")
		return nil, err
	}

	// etcd 연결
	err = etcd.Initialize()
	if err != nil {
		logrus.Error("Failed to initialize etcd")
		return nil, err
	}

	manager.metricL = &sync.RWMutex{}

	manager.AgentQueueTTL = map[string]time.Time{}
	manager.AgentQueueColN = map[string]int{}

	return &manager, nil
}

// 기존의 실시간 모니터링 데이터 삭제
func (manager *CollectManager) FlushMonitoringData() error {
	// 모니터링 콜렉터 태그 정보 삭제
	etcd.GetInstance().DeleteMetric("/collector")

	// 실시간 모니터링 정보 삭제
	etcd.GetInstance().DeleteMetric("/vm")
	etcd.GetInstance().DeleteMetric("/mon")

	manager.SetConfigurationToETCD()

	return nil
}

func (manager *CollectManager) SetConfigurationToETCD() error {
	var monConfigMap map[string]interface{}
	err := mapstructure.Decode(config.GetInstance().Monitoring, &monConfigMap)
	if err != nil {
		return err
	}

	// etcd 저장소에 모니터링 정책 저장 후 결과 값 반환
	err = etcd.GetInstance().WriteMetric("/mon/config", monConfigMap)
	if err != nil {
		return err
	}
	return nil
}

func (manager *CollectManager) CreateLoadBalancer(wg *sync.WaitGroup) error {

	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", config.GetInstance().CollectManager.CollectorPort))
	if err != nil {
		logrus.Error("Failed to resolve UDP server address: ", err)
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logrus.Error("Failed to listen UDP server address: ", err)
		return err
	}
	//manager.WaitGroup = wg
	manager.UdpCOnn = udpConn

	manager.WaitGroup.Add(1)
	go manager.ManageAgentTtl(manager.WaitGroup)

	manager.WaitGroup.Add(1)
	go manager.StartLoadBalancer(manager.UdpCOnn, manager.WaitGroup)

	return nil
}

func (manager *CollectManager) StartLoadBalancer(udpConn net.PacketConn, wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		metric := collector.TelegrafMetric{}
		buf := make([]byte, 1024*10)

		n, _, err := udpConn.ReadFrom(buf)

		if err != nil {
			logrus.Error("UDPLoadBalancer : failed to read bytes: ", err)
		}
		manager.metricL.Lock()
		if err := json.Unmarshal(buf[0:n], &metric); err != nil {
			logrus.Error("Failed to decode json to buf: ", string(buf[0:n]))
			continue
		}
		manager.metricL.Unlock()
		vmId := metric.Tags["vmId"].(string)

		manager.AgentQueueTTL[vmId] = time.Now()

		_, alreadyRegistered := manager.AgentQueueColN[vmId]

		if !alreadyRegistered {
			manager.metricL.Lock()
			manager.AgentQueueColN[vmId] = -1
			manager.metricL.Unlock()
		}

		err = manager.ManageAgentQueue(vmId, manager.AgentQueueColN, metric)
		if err != nil {
			logrus.Error("ManageAgentQueue Error", err)
		}
	}
}

func (manager *CollectManager) ManageAgentQueue(vmId string, AgentQueueColN map[string]int, metric collector.TelegrafMetric) error {
	colN := AgentQueueColN[vmId]
	colUUID := ""
	var cashingCAddr *collector.MetricCollector
	// Case : new Data which is not allocated at collector
	for idx, cUUID := range manager.CollectorIdx {

		cAddr := manager.CollectorUUIDAddr[cUUID]

		if cAddr != nil {
			if _, alreadyRegistered := (*cAddr).MarkingAgent[vmId]; alreadyRegistered {
				if idx != 0 {
					cashingCAddr = cAddr
					break
				}
				colUUID = manager.CollectorIdx[colN]
				*(manager.TransmitDataChan[colUUID]) <- metric
				return nil
			}
		}
	}

	for idx, cUUID := range manager.CollectorIdx {

		cAddr := manager.CollectorUUIDAddr[cUUID]

		if len((*cAddr).MarkingAgent) < config.GetInstance().Monitoring.MaxHostCount {

			if cashingCAddr != nil {
				delete((*cashingCAddr).MarkingAgent, vmId)
			}
			manager.metricL.Lock()
			(*cAddr).MarkingAgent[vmId] = vmId
			manager.metricL.Unlock()
			AgentQueueColN[vmId] = idx
			colN = AgentQueueColN[vmId]
			colUUID = manager.CollectorIdx[colN]
			*(manager.TransmitDataChan[colUUID]) <- metric
			return nil
		}
	}

	return nil
}

func (manager *CollectManager) ManageAgentTtl(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		currentTime := time.Now()
		if len(manager.AgentQueueTTL) != 0 {
			manager.metricL.RLock()
			for vmId, arrivedTime := range manager.AgentQueueTTL {

				if currentTime.Sub(arrivedTime) > time.Duration(config.GetInstance().Monitoring.AgentTTL)*time.Second {
					if _, ok := manager.AgentQueueTTL[vmId]; ok {
						//manager.metricL.RLock()
						delete(manager.AgentQueueTTL, vmId)
						//manager.metricL.RUnlock()
					}
					colN := manager.AgentQueueColN[vmId]
					cUUID := ""
					if colN >= 0 && colN < len(manager.CollectorIdx) {
						cUUID = manager.CollectorIdx[colN]
					} else {
						continue
					}
					c := manager.CollectorUUIDAddr[cUUID]
					if _, ok := manager.AgentQueueColN[vmId]; ok {
						delete(manager.AgentQueueColN, vmId)
					}
					if _, ok := (*c).MarkingAgent[vmId]; ok {
						delete((*c).MarkingAgent, vmId)
					}
					err := etcd.GetInstance().DeleteMetric(fmt.Sprintf("/collector/%s/vm/%s", cUUID, vmId))
					if err != nil {
						logrus.Error("Fail to delete vmInfo ETCD data")
					}
					err = etcd.GetInstance().DeleteMetric(fmt.Sprintf("/vm/%s", vmId))
					if err != nil {
						logrus.Error("Fail to delete expired ETCD data")
					}
				}
			}
			manager.metricL.RUnlock()
		}
		time.Sleep(1 * time.Second)
	}
}

//func (manager *CollectManager) StartCollector(wg *sync.WaitGroup, aggregateChan chan string) error {
func (manager *CollectManager) StartCollector(wg *sync.WaitGroup) error {
	manager.WaitGroup = wg

	manager.CollectorIdx = []string{}
	manager.CollectorUUIDAddr = map[string]*collector.MetricCollector{}
	manager.AggregatingChan = map[string]*chan string{}
	manager.TransmitDataChan = map[string]*chan collector.TelegrafMetric{}

	for i := 0; i < config.GetInstance().CollectManager.CollectorCnt; i++ {
		err := manager.CreateCollector()
		if err != nil {
			logrus.Error("failed to create collector", err)
			continue
		}
	}

	return nil
}

func (manager *CollectManager) CreateCollector() error {
	// 실시간 데이터 저장을 위한 collector 고루틴 실행
	mc := collector.NewMetricCollector(map[string]string{}, manager.metricL, config.GetInstance().Monitoring.CollectorInterval, collector.AVG, manager.AggregatingChan, manager.TransmitDataChan)
	manager.metricL.Lock()
	manager.CollectorIdx = append(manager.CollectorIdx, mc.UUID)
	manager.metricL.Unlock()
	transmitDataChan := make(chan collector.TelegrafMetric)
	manager.WaitGroup.Add(1)
	go mc.StartCollector(manager.UdpCOnn, manager.WaitGroup, &transmitDataChan)

	manager.WaitGroup.Add(1)
	aggregateChan := make(chan string)
	go mc.StartAggregator(manager.WaitGroup, &aggregateChan)

	manager.TransmitDataChan[mc.UUID] = &transmitDataChan
	manager.CollectorUUIDAddr[mc.UUID] = &mc
	manager.AggregatingChan[mc.UUID] = &aggregateChan

	return nil
}

func (manager *CollectManager) StopCollector(uuid string) error {
	if _, ok := manager.CollectorUUIDAddr[uuid]; ok {
		// 실행 중인 콜렉터 고루틴 종료 (콜렉터 활성화 플래그 변경)
		manager.CollectorUUIDAddr[uuid].Active = false
		delete(manager.CollectorUUIDAddr, uuid)
		//*(manager.TransmitDataChan[uuid]) <- collector.TelegrafMetric{}
		//manager.CollectorCnt -= 1
		return nil
	} else {
		return errors.New(fmt.Sprintf("failed to get collector by id, uuid: %s", uuid))
	}
}

//func (manager *CollectManager) GetConfigInfo() (Monitoring, error) {
//	// etcd 저장소 조회
//	configNode, err := manager.Etcd.ReadMetric("/mon/config")
//	if err != nil {
//		return Monitoring{}, err
//	}
//	// Monitoring 매핑
//	var config Monitoring
//	err = json.Unmarshal([]byte(configNode.Value), &config)
//	if err != nil {
//		return Monitoring{}, err
//	}
//	return config, nil
//}

func (manager *CollectManager) StartAggregateScheduler(wg *sync.WaitGroup, c *map[string]*chan string) {
	defer wg.Done()
	for {
		// aggregate 주기 정보 조회
		collectorInterval := config.GetInstance().Monitoring.CollectorInterval

		//// Print Session Start /////
		//fmt.Print("\nTTL queue List : ")
		//sortedAgentQueueTTL := make([] int, 0)
		//for key, _ := range manager.AgentQueueTTL{
		//	value, _ := strconv.Atoi(strings.Split(key,"-")[2])
		//	sortedAgentQueueTTL = append(sortedAgentQueueTTL, value)
		//}
		//sort.Slice(sortedAgentQueueTTL, func(i, j int) bool {
		//	return sortedAgentQueueTTL[i] < sortedAgentQueueTTL[j]
		//})
		//for _, value := range sortedAgentQueueTTL {
		//	fmt.Print(value, ", ")
		//}
		//fmt.Print(fmt.Sprintf(" / Total : %d", len(sortedAgentQueueTTL)))
		//fmt.Print("\n")
		//fmt.Println("The number of collector : ", len(manager.CollectorIdx))
		//// Print Session End /////

		time.Sleep(time.Duration(collectorInterval) * time.Second)

		for _, channel := range *c {
			*channel <- "aggregate"
		}
	}
}

// 콜렉터 스케일 인/아웃 관리 스케줄러
func (manager *CollectManager) StartScaleScheduler(wg *sync.WaitGroup) {
	defer wg.Done()
	cs := NewCollectorScheduler(manager)
	for {
		// 스케줄링 주기 정보 조회
		schedulingInterval := config.GetInstance().Monitoring.SchedulingInterval

		time.Sleep(time.Duration(schedulingInterval) * time.Second)

		// Check Scale-In/Out Logic (len(AgentTTLQueue) 기준 Scaling In/Out)
		err := cs.CheckScaleCondition()
		if err != nil {
			logrus.Error("failed to check scale in/out condition", err)
		}
	}
}
