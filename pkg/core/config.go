package core

import (
	"sync"

	metricstore "github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore"
)

type Core struct {
	Etcd     realtimestore.Storage
	InfluxDB metricstore.Storage
}

var CoreConfig Core
var once sync.Once

func InitCoreConfig(etcd realtimestore.Storage) {
	once.Do(func() {
		CoreConfig = Core{
			Etcd: etcd,
		}
	})
}
