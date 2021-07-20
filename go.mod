module github.com/cloud-barista/cb-dragonfly

go 1.15

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.3
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/Scalingo/go-utils v7.1.0+incompatible
	github.com/bramvdbogaerde/go-scp v1.0.0
	github.com/cloud-barista/cb-log v0.4.0
	github.com/cloud-barista/cb-spider v0.4.5
	github.com/cloud-barista/cb-store v0.4.1
	github.com/confluentinc/confluent-kafka-go v1.7.0 // indirect
	github.com/coreos/bbolt v1.3.4 // indirect
	github.com/deepmap/oapi-codegen v1.8.1 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/influxdata/influxdb v1.9.2 // indirect
	github.com/influxdata/influxdb-client-go v1.4.0
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/influxdata/line-protocol v0.0.0-20210311194329-9aa0e372d097 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/labstack/echo/v4 v4.3.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/shaodan/kapacitor-client v0.0.0-20181228024026-84c816949946
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/assertions v1.1.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	google.golang.org/grpc v1.39.0
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.7.0
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/yaml v1.2.0 // indirect
)
