package etcd

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client"
	"sync"
)

type ClientOptions struct {
	Endpoints string
}

type Config struct {
	ClientOptions ClientOptions
}

type Storage struct {
	Config Config
	Client client.Client
	L 	 *sync.RWMutex
}

func (s *Storage) Init() error {
	cfg := client.Config{
		Endpoints: []string{
			s.Config.ClientOptions.Endpoints,
		},
	}
	if client, err := client.New(cfg); err != nil {
		logrus.Error(err)
		return err
	} else {
		s.Client = client
	}
	s.L = &sync.RWMutex{}
	return nil
}

//func (s *Storage) WriteMetric(key string, metric map[string]interface{}) error {
func (s *Storage) WriteMetric(key string, metric interface{}) error {
	//s.L.Lock()
	kapi := client.NewKeysAPI(s.Client)

	//fmt.Println(fmt.Sprintf("[ETCD] receive &metric : %p", &metric))

	metricVal := metric.(string)
	//var metricVal string
	//
	//_, ok := metric.(map[string]interface{})
	//if ok {
	//
	//	s.L.Lock()
	//	bytes, err := json.Marshal(metric)
	//	s.L.Unlock()
	//
	//	if err != nil {
	//		logrus.Error("Failed to marshaling realtime monitoring data to JSON: ", err)
	//		return err
	//	}
	//
	//	metricVal = fmt.Sprintf("%s", bytes)
	//} else {
	//	metricVal = metric.(string)
	//}

	// 실시간 모니터링 데이터 저장
	// TODO: 추후 모니터링 데이터 TTL(Time To Live) 설정 추가
	opts := client.SetOptions{TTL: -1}

	s.L.RLock()
	_, err := kapi.Set(context.Background(), key, fmt.Sprintf("%s", metricVal), &opts)
	s.L.RUnlock()
	if err != nil {
		logrus.Error("Failed to write realtime monitoring data to ETCD : ", err)
	//	s.L.Unlock()
		return err
	}
	//s.L.Unlock()
	//logrus.Debug("Write is done. Response is %q\n", resp)
	return nil
}

//func (s *Storage) ReadMetric(key string) (map[string]interface{}, error) {
func (s *Storage) ReadMetric(key string) (*client.Node, error) {
	//s.L.RLock()
	kapi := client.NewKeysAPI(s.Client)
	// fmt.Println(" ETCD Key : ",key)
	// 실시간 모니터링 데이터 조회
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		logrus.Error("Failed to read realtime monitoring data to ETCD : ", err)
		//s.L.RUnlock()
		return nil, err
	}

	// 실시간 모니터링 데이터 파싱
	// TODO: 추후 etcd Node 형태가 아닌 별도 구조체 형태로 데이터 파싱 필요
	/*for i, node := range nodeArr {
		fmt.Println(i)
		var metric map[string]types.CPU
		if err := json.Unmarshal([]byte(node.Value), &metric); err != nil {
			logrus.Error(err)
			continue
		}
		//spew.Dump(metric)
	}*/

	if resp == nil {
	//	s.L.RUnlock()
		return nil, nil
	}

	//s.L.RUnlock()
	//logrus.Debug("Read is done. Response is %q\n", resp)
	return resp.Node, nil
}

func (s *Storage) DeleteMetric(key string) error {
	kapi := client.NewKeysAPI(s.Client)

	// 실시간 모니터링 데이터 삭제
	opts := client.DeleteOptions{Recursive: true}
	//resp, err := kapi.Delete(context.Background(), key, &opts)
	_, err := kapi.Delete(context.Background(), key, &opts)
	if err != nil {
		logrus.Error("Failed to delete realtime monitoring data to ETCD : ", err)
		return err
	}

	//logrus.Debug("Delete is done. Response is %q\n", resp)
	return nil
}