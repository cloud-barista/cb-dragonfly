package alert

import (
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"strings"
	"time"

	kclient "github.com/shaodan/kapacitor-client"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
)

const (
	kapacitorTimeout = 5 * time.Minute
)

//var once sync.Once
//var client *kclient.Client

func newClient() (*kclient.Client, error) {
	var kapacitorPort int
	if strings.EqualFold(config.GetDefaultConfig().GetMonConfig().DeployType, "local") {
		kapacitorPort = config.GetInstance().GetKapacitorConfig().ExternalPort // 29092
	} else {
		kapacitorPort = config.GetInstance().GetKapacitorConfig().InternalPort // 9092
	}
	kapacitorConfig := kclient.Config{
		URL:                fmt.Sprintf("%s:%d", config.GetDefaultConfig().GetKapacitorConfig().GetKapacitorEndpointUrl(), kapacitorPort),
		Timeout:            time.Duration(kapacitorTimeout),
		InsecureSkipVerify: true,
	}
	client, err := kclient.New(kapacitorConfig)
	if client != nil {
		if _, _, err := client.Ping(); err != nil {
			util.GetLogger().Error(fmt.Sprintf("failed to ping kapacitor, error=%s", err))
		}
	}
	return client, err
}

func GetClient() *kclient.Client {
	c, _ := newClient()
	return c
}
