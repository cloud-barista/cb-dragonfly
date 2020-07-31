package core

import (
	"sync"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	metricstore "github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore"
)

type Core struct {
	Config   config.Config
	Etcd     realtimestore.Storage
	InfluxDB metricstore.Storage
}

var CoreConfig Core
var once sync.Once

func InitCoreConfig(config config.Config, influxDB metricstore.Storage, etcd realtimestore.Storage) {
	once.Do(func() {
		CoreConfig = Core{
			Config:   config,
			Etcd:     etcd,
			InfluxDB: influxDB,
		}
	})
}
