module github.com/cloud-barista/cb-dragonfly

go 1.12

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.3

require (
	github.com/Scalingo/go-utils v5.5.14+incompatible
	github.com/bramvdbogaerde/go-scp v0.0.0-20200119201711-987556b8bdd7
	github.com/cloud-barista/cb-log v0.2.0-cappuccino.0.20200913031717-ff545833c178 // indirect
	github.com/cloud-barista/cb-spider v0.2.0-cappuccino.0.20200925073009-73c399c7f818
	github.com/cloud-barista/cb-store v0.2.0-cappuccino.0.20200924125209-c313bd2a3987 // indirect
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/google/uuid v1.1.1
	github.com/influxdata/influxdb v1.7.8
	github.com/influxdata/influxdb-client-go v0.0.1
	github.com/influxdata/influxdb1-client v0.0.0-20190809212627-fc22c7df067e
	github.com/labstack/echo/v4 v4.1.10
	github.com/mitchellh/mapstructure v1.3.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/snowzach/rotatefilehook v0.0.0-20180327172521-2f64f265f58c // indirect
	github.com/spf13/viper v1.7.1
	go.etcd.io/etcd v3.3.18+incompatible
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)
