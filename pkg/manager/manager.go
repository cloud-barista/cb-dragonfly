package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore/etcd"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"
)

// TODO: implements
// TODO: 1. API Server
// TODO: 2. Scheduling Collector...
// TODO: 3. Configuring Policy...

type CollectManager struct {
	Config            Config
	InfluxdDB         metricstore.Storage
	Etcd              realtimestore.Storage
	Aggregator        collector.Aggregator
	WaitGroup         *sync.WaitGroup
	UdpCOnn           *net.UDPConn
	CollectorIdx      []string
	CollectorUUIDAddr map[string]*collector.MetricCollector
	AggregatingChan   map[string]*chan string
	TransmitDataChan  map[string]*chan collector.TelegrafMetric
	AgentQueueTTL     map[string]time.Time
	AgentQueueColN    map[string]int
	//HostInfo      collector.HostInfo
	//HostCnt       int
}

// 콜렉터 매니저 초기화
func NewCollectorManager() (*CollectManager, error) {
	manager := CollectManager{}
	err := manager.LoadConfiguration()
	if err != nil {
		return nil, err
	}

	influxConfig := influxdbv1.Config{
		ClientOptions: []influxdbv1.ClientOptions{
			{
				URL:      manager.Config.InfluxDB.EndpointUrl,
				Username: manager.Config.InfluxDB.UserName,
				Password: manager.Config.InfluxDB.Password,
			},
		},
		Database: manager.Config.InfluxDB.Database,
	}

	// InfluxDB 연결
	influx, err := metricstore.NewStorage(metricstore.InfluxDBV1Type, influxConfig)
	if err != nil {
		logrus.Error("Failed to initialize influxDB")
		return nil, err
	}
	manager.InfluxdDB = influx

	etcdConfig := etcd.Config{
		ClientOptions: etcd.ClientOptions{
			Endpoints: manager.Config.Etcd.EndpointUrl,
		},
	}

	// etcd 연결
	etcd, err := realtimestore.NewStorage(realtimestore.ETCDV2Type, etcdConfig)
	if err != nil {
		logrus.Error("Failed to initialize etcd")
		return nil, err
	}
	manager.Etcd = etcd

	manager.AgentQueueTTL = map[string]time.Time{}
	manager.AgentQueueColN = map[string]int{}

	return &manager, nil
}

// 기존의 실시간 모니터링 데이터 삭제
func (manager *CollectManager) FlushMonitoringData() error {
	// 모니터링 콜렉터 태그 정보 삭제
	//manager.Etcd.DeleteMetric("/host-list")
	manager.Etcd.DeleteMetric("/collector")

	// 실시간 모니터링 정보 삭제
	manager.Etcd.DeleteMetric("/host")

	return nil
}

// config 파일 로드
func (manager *CollectManager) LoadConfiguration() error {
	configPath := os.Getenv("CBMON_PATH") + "/conf/config.yaml"

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Error("Failed to read configuration file in: ", configPath)
		return err
	}

	err = yaml.Unmarshal(bytes, &manager.Config)
	if err != nil {
		logrus.Error("Failed to unmarshal configuration file")
		return err
	}

	return nil
}

// TODO: 모니터링 정책 설정
func (manager *CollectManager) SetConfiguration() error {
	return nil
}

func (manager *CollectManager) CreateLoadBalancer(wg *sync.WaitGroup) error {

	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", manager.Config.CollectManager.CollectorPort))
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
	metric := collector.TelegrafMetric{}
	//logicStartTime := time.Now()
	monConfig, err := manager.GetConfigInfo()
	if err != nil {
		logrus.Error("Fail to get monConfig Info")
	}

	for {

		buf := make([]byte, 1024*10)

		n, _, err := udpConn.ReadFrom(buf)

		if err != nil {
			logrus.Error("UDPLoadBalancer : failed to read bytes: ", err)
		}
		if err := json.Unmarshal(buf[0:n], &metric); err != nil {
			logrus.Error("Failed to decode json to buf: ", string(buf[0:n]))
			continue
		}
		//fmt.Println("\n metric : ", metric)
		hostId := metric.Tags["hostID"].(string)

		manager.AgentQueueTTL[hostId] = time.Now()
		manager.AgentQueueColN[hostId] = 0
		/*
				if len(manager.CollectorIdx)*monConfig.MaxHostCount < len((manager.AgentQueueTTL)) {
					continue
					//manager.CollectorQueue[manager.ManageCollectorQueue(metric)][hostID] = metric
				}else{
					manager.ManageAgentQueue(hostId, manager.AgentQueueColN, metric)
				}
				if len(manager.AgentQueueTTL) == 0{ // first executing logic
					manager.AgentQueueTTL[hostId] = time.Now()
					manager.AgentQueueColN[hostId] = 0
			//		fmt.Println("AgentQueueTTL == 0")
				}else{
					// if new hostId
					if _, ok := (manager.AgentQueueTTL)[hostId]; !ok{
						// input new hostId at TTL queue
						(manager.AgentQueueTTL)[hostId] = time.Now()
						// input new hostId at ColN queue, default ColN is 0
						manager.AgentQueueColN[hostId] = 0
					}else{ // if exist hostId
						currentTime := time.Now()
						fmt.Println("TTL : ",currentTime.Sub((manager.AgentQueueTTL)[hostId]))
						fmt.Println("SETtime : ",time.Duration(monConfig.AgentTtl)*time.Second)
						// check TTL. if Live time is more than set Live time, than delete data
						if currentTime.Sub((manager.AgentQueueTTL)[hostId]) > time.Duration(monConfig.AgentTtl)*time.Second {

							colN := manager.AgentQueueColN[hostId]
							cUUID := manager.CollectorIdx[colN]
							c := manager.CollectorUUIDAddr[cUUID]

							delete((manager.AgentQueueTTL), hostId)
							delete(manager.AgentQueueColN, hostId)
							delete((*c).MarkingAgent, hostId)
							// add delete logic

						}else { // inner TTL data, update current time
							(manager.AgentQueueTTL)[hostId] = time.Now()
						}
					}
				}*/
		// drop metric section (before scaleScheduling, it will drop the input data)
		if len(manager.CollectorIdx)*monConfig.MaxHostCount < len(manager.AgentQueueTTL) {
			continue
			//manager.CollectorQueue[manager.ManageCollectorQueue(metric)][hostID] = metric
		} else {
			manager.ManageAgentQueue(hostId, manager.AgentQueueColN, metric)
		}
	}
}

func (manager *CollectManager) ManageAgentTtl(wg *sync.WaitGroup) {

	defer wg.Done()

	monConfig, err := manager.GetConfigInfo()
	if err != nil {
		logrus.Error("Fail to get monConfig Info")
	}

	for {
		currentTime := time.Now()
		if len(manager.AgentQueueTTL) != 0 {
			for hostId, arrivedTime := range manager.AgentQueueTTL {

				if currentTime.Sub(arrivedTime) > time.Duration(monConfig.AgentTtl)*time.Second {

					if _, ok := manager.AgentQueueTTL[hostId]; ok {
						delete(manager.AgentQueueTTL, hostId)
					}
					if _, ok := manager.AgentQueueColN[hostId]; ok {
						delete(manager.AgentQueueColN, hostId)
					}
					colN := manager.AgentQueueColN[hostId]
					cUUID := manager.CollectorIdx[colN]
					c := manager.CollectorUUIDAddr[cUUID]
					if _, ok := (*c).MarkingAgent[hostId]; ok {
						delete((*c).MarkingAgent, hostId)
					}
				}

			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (manager *CollectManager) ManageAgentQueue(hostId string, AgentQueueColN map[string]int, metric collector.TelegrafMetric) error {

	colN := AgentQueueColN[hostId]
	colUUID := ""
	// Case : new Data which is not allocated at collector
	if colN == 0 {

		for idx, cUUID := range manager.CollectorIdx {

			cAddr := manager.CollectorUUIDAddr[cUUID]

			if _, alreadyRegistered := (*cAddr).MarkingAgent[hostId]; alreadyRegistered {
				break
			}

			if len((*cAddr).MarkingAgent[hostId]) < manager.Config.Monitoring.MaxHostCount {

				if _, alreadyRegistered := (*cAddr).MarkingAgent[hostId]; !alreadyRegistered {

					(*cAddr).MarkingAgent[hostId] = hostId
					AgentQueueColN[hostId] = idx
					colN = AgentQueueColN[hostId]

					break
				}
			}
		}
		// sending Metric to collector
		colUUID = manager.CollectorIdx[colN]
		*(manager.TransmitDataChan[colUUID]) <- metric

	} else {
		for idx, cUUID := range manager.CollectorIdx {
			cAddr := manager.CollectorUUIDAddr[cUUID]

			if _, alreadyRegistered := (*cAddr).MarkingAgent[hostId]; alreadyRegistered {
				break
			}

			if len((*cAddr).MarkingAgent) < manager.Config.Monitoring.MaxHostCount {

				(*cAddr).MarkingAgent[hostId] = hostId

				origin_cUUID := manager.CollectorIdx[colN]
				origtin_cUUIDAddr := manager.CollectorUUIDAddr[origin_cUUID]
				delete((*origtin_cUUIDAddr).MarkingAgent, hostId)

				AgentQueueColN[hostId] = idx
				colN = AgentQueueColN[hostId]

				break
			}
		}

		// sending Metric to collector
		colUUID = manager.CollectorIdx[colN]
		*(manager.TransmitDataChan[colUUID]) <- metric
	}
	return nil
}

//func (manager *CollectManager) StartCollector(wg *sync.WaitGroup, aggregateChan chan string) error {
func (manager *CollectManager) StartCollector(wg *sync.WaitGroup) error {
	manager.WaitGroup = wg

	manager.CollectorIdx = []string{}
	manager.CollectorUUIDAddr = map[string]*collector.MetricCollector{}
	manager.AggregatingChan = map[string]*chan string{}
	manager.TransmitDataChan = map[string]*chan collector.TelegrafMetric{}

	for i := 0; i < manager.Config.CollectManager.CollectorCnt; i++ {
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
	mc := collector.NewMetricCollector(map[string]string{}, manager.Config.Monitoring.CollectorInterval, &manager.Etcd, &manager.InfluxdDB, collector.AVG, manager.AggregatingChan, manager.TransmitDataChan)

	manager.CollectorIdx = append(manager.CollectorIdx, mc.UUID)

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

		*(manager.TransmitDataChan[uuid]) <- collector.TelegrafMetric{}
		//manager.CollectorCnt -= 1
		return nil
	} else {
		return errors.New(fmt.Sprintf("failed to get collector by id, uuid: %s", uuid))
	}
}

func (manager *CollectManager) GetConfigInfo() (MonConfig, error) {
	// etcd 저장소 조회
	configNode, err := manager.Etcd.ReadMetric("/mon/config")
	if err != nil {
		return MonConfig{}, err
	}
	// MonConfig 매핑
	var config MonConfig
	err = json.Unmarshal([]byte(configNode.Value), &config)
	if err != nil {
		return MonConfig{}, err
	}
	return config, nil
}

// Need collector labels

func (manager *CollectManager) StartAggregateScheduler(wg *sync.WaitGroup, c *map[string]*chan string) {
	defer wg.Done()
	for {
		/*select {
		case <-ctx.Done():
			logrus.Debug("Stop scheduling for aggregate metric")
			return
		default:
		}*/
		// aggregate 주기 정보 조회
		monConfig, err := manager.GetConfigInfo()
		if err != nil {
			logrus.Error("failed to get monitoring config info", err)
		}
		/*
			fmt.Print("\nmanager.AgentQueueColN : ")
			fmt.Print("[0] : ")
			for key, val := range manager.AgentQueueColN{
				if val == 0 {
					fmt.Print(key,", ")
				}
			}
			fmt.Println("")
			fmt.Print("manager.AgentQueueTTL : ")
			for key, _ := range manager.AgentQueueTTL{
				fmt.Print(key, ", ")
			}
			fmt.Print("\n")
			fmt.Println("manager.CollectorIdx : ", manager.CollectorIdx)
			fmt.Println("manager.CollectorUUIDAddr : ", manager.CollectorUUIDAddr)*/
		time.Sleep(time.Duration(monConfig.CollectorInterval) * time.Second)
		//manager.HostCnt = len(*manager.HostInfo.HostMap)

		for _, channel := range *c {

			//fmt.Println("StartScheduler Current Channel : ",*c)
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
		monConfig, err := manager.GetConfigInfo()
		if err != nil {
			logrus.Error("failed to get monitoring config info", err)
		}

		time.Sleep(time.Duration(monConfig.SchedulingInterval) * time.Second)

		// Check Scale-In/Out Logic (호스트 수 기준 Scaling In/Out)
		err = cs.CheckScaleCondition()
		if err != nil {
			logrus.Error("failed to check scale in/out condition", err)
		}
	}
}
