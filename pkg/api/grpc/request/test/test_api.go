package main

import (
	"fmt"
	"time"

	"github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/logger"
	api "github.com/cloud-barista/cb-dragonfly/pkg/api/grpc/request"
)

// 테스트 환경에 맞게 파라미터 수정 필요
var (
	nsId   string = "NS-01"
	mcisId string = "openstack-config01-developer"
	vmId   string = "openstack-config01-developer-01"

	sshKey string = `-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAvajQVR3JsM5cX8wWwi4FS9ZC6ecQgHK21UklUtrxPPBlFfPq\namxbFql2y9FTL2NM9ugGXxj+5QxDVRrDcS84i/4kgFyA7mVOjwHbuVeZKQAODwvU\ninuHz6JBYoxX3Fs99i7RB6TjM1pYiMmg8z/Gy9ve4CH+yRasOB3glXwLL0XfLEED\n8Iw0PtlQKxtXcO9kCkttDkFD/ce8RU2xcZp9keNI9eNTIr4xwi/yhd9ww/fOqilX\nH1obVIrnD9Fr1xu5vvSj5AKtpFNcC9I1uz497tFaIqJBNwYL8/sS2pR7rSdeNpCG\nnxh3nTlWiLhcqTDyit1hL70rUM5BsQKgETQuUQIDAQABAoIBAEfeKmO2j/EBoZtj\neNRIIBWmsWB1AJnL3mBgAVauRG+1IHj7Hr8JJFMoEC4Xug/g7w84yQFMNXqR9QnQ\nxHKlVCYoPaiuZOTxWp1yNNK80PrqXGlzMCzxtnsFnwU67ShBIu+gufDNmJKjD511\n2hmS9z/Up1YDS8rjXos9NxcuFAbrKjAHfIm3leHW497sqmdzFEfuwGG1Nh8Wn/1/\nqhvE+SH2GUvsmMhL2BzvBLvMn+FcgU/xzA7V8xUzZ/h2xDve77uWJe8Zt4aCZ3ts\nA00GwNxDK9BLZNvxrI1L9NQsMMFp6SbXMwUJQjUraZxXEx8YNU5SuX46HjL916JG\nUqB12mkCgYEA3jtcOwxl+PqfVB8PkJ5nRLbNOSfS6lBMGiNtTHSeWja17FQ/t20+\n4fenYc0GrvfvELWpqKugnxNkE/XQOfLWefMzyx1M5XTDMSs+q/qDLxqBC2uVY9sx\niSo+N472Abb/MWdCDDEt+i7rmIp3jf4CYErock5Y+7xjYJR7x/Js6bsCgYEA2npp\naEnkd73mQygZnhUMEsKQ8LR1u6AdkWPf8ekZXRLm+8CbVWVqMh2LyPTefefc7JX6\nvmHkD2KrcOeWuNfjZ87TkmXbAnepQiKI9EUHXLrp0fJ0qKszHeECSwRbRsDVs52t\nld42nYbnTzNVqBJ1d/p6Pm5k8bz9uh+pNAr7MWMCgYBNy33n9dkkpads7Uqnl6wS\n8+M3pOdCu0VIySoT36cncYuR5ZRAg+/FbsqbhAhY69Y2hUGVGC+sQD+CdUSlZIsM\nOcThz6oBkTRbXAYech3GOYL/GnQ7dpoKqE0LafJYe1UuWDVYy0aLFC3JQn/Dpy3x\nw4dHrIGd7j7jqlcCkazqzQKBgFTAbclt+LnWqc9da+qeAYv5fB5T8uPw7edrlgES\nyqsHXSFGCzvqVnLQcVxoWTMAUfxZBeb6tGyfeyAWRqq33Nh1LC+7YWUopDkqinQv\nnvaC75do9YZEu1SY57nQG9Rrk7rUrPTZOdiL74kSweSsHHOJcAht7Ky2ArtD8vBk\nXiM/AoGBAIuBYp6fNBmKureQBuVmHZzOKuNJCNn1Bnfh+Ok9J4FQW9TTddLd9xee\nFXKkqum3nqAY9HxXydX8kdpRsbfjbKB7ruC8pSVpaoNpn/UwVc3Cv6wma+GU6BIC\nggdUdyWZ92/bd1C8VOoyn+bAGTdTqGSxUivIvdL5sPWH9/YdUMaQ\n-----END RSA PRIVATE KEY-----\n`
	vmIp   string = "192.168.201.36"
	vmUser string = "ubuntu"
)

func main() {
	SimpleMONApiTest()
	ConfigMONApiTest()

	GetVMMonInfoApiTest()
	GetVMLatestMonInfoApiTest()
	MonConfigApiTest()
	InstallTelegrafApiTest()
}

// SimpleMONApiTest - 환경설정함수를 이용한 간단한 MON API 호출
func SimpleMONApiTest() {

	fmt.Print("\n\n============= SimpleMONApiTest() =============\n")

	logger := logger.NewLogger()

	mon := api.NewMONManager()

	err := mon.SetServerAddr("localhost:50253")
	if err != nil {
		logger.Fatal(err)
	}

	err = mon.SetTimeout(90 * time.Second)
	if err != nil {
		logger.Fatal(err)
	}

	/* 서버가 TLS 가 설정된 경우
	err = mon.SetTLSCA(os.Getenv("CBMON_ROOT") + "/certs/ca.crt")
	if err != nil {
		logger.Fatal(err)
	}
	*/

	/* 서버가 JWT 인증이 설정된 경우
	err = mon.SetJWTToken("xxxxxxxxxxxxxxxxxxx")
	if err != nil {
		logger.Fatal(err)
	}
	*/

	err = mon.Open()
	if err != nil {
		logger.Fatal(err)
	}

	result, err := mon.GetVMMonCpuInfoByParam(nsId, mcisId, vmId, "h", "avg", "1h")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	mon.Close()
}

// ConfigMONApiTest - 환경설정파일을 이용한 MON API 호출
func ConfigMONApiTest() {

	fmt.Print("\n\n============= ConfigMONApiTest() =============\n")

	logger := logger.NewLogger()

	mon := api.NewMONManager()

	err := mon.SetConfigPath("./grpc_conf.yaml")
	if err != nil {
		logger.Fatal(err)
	}

	err = mon.Open()
	if err != nil {
		logger.Fatal(err)
	}

	result, err := mon.GetVMMonCpuInfoByParam(nsId, mcisId, vmId, "h", "avg", "1h")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	mon.Close()
}

// GetVMMonInfoApiTest - GetVMMonInfo 관련 API 호출 테스트
func GetVMMonInfoApiTest() {

	fmt.Print("\n\n============= GetVMMonInfoApiTest() =============\n")

	logger := logger.NewLogger()

	mon := api.NewMONManager()

	err := mon.SetConfigPath("./grpc_conf.yaml")
	if err != nil {
		logger.Fatal(err)
	}

	err = mon.Open()
	if err != nil {
		logger.Fatal(err)
	}

	result, err := mon.GetVMMonCpuInfoByParam(nsId, mcisId, vmId, "h", "avg", "1h")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	/*
		result, err = mon.GetVMMonCpuFreqInfoByParam(nsId, mcisId, vmId, "h", "avg", "1h")
		if err != nil {
			logger.Fatal(err)
		}

		fmt.Printf("\nresult :\n%s\n", result)
	*/

	result, err = mon.GetVMMonMemoryInfoByParam(nsId, mcisId, vmId, "h", "avg", "1h")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	result, err = mon.GetVMMonDiskInfoByParam(nsId, mcisId, vmId, "h", "avg", "1h")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	result, err = mon.GetVMMonNetworkInfoByParam(nsId, mcisId, vmId, "h", "avg", "1h")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	mon.Close()
}

// GetVMLatestMonInfoApiTest - GetVMLatestMonInfo 관련 API 호출 테스트
func GetVMLatestMonInfoApiTest() {

	fmt.Print("\n\n============= GetVMLatestMonInfoApiTest() =============\n")

	logger := logger.NewLogger()

	mon := api.NewMONManager()

	err := mon.SetConfigPath("./grpc_conf.yaml")
	if err != nil {
		logger.Fatal(err)
	}

	err = mon.Open()
	if err != nil {
		logger.Fatal(err)
	}

	result, err := mon.GetVMLatestMonCpuInfoByParam(nsId, mcisId, vmId, "avg")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	/*
		result, err = mon.GetVMLatestMonCpuFreqInfoByParam(nsId, mcisId, vmId, "avg")
		if err != nil {
			logger.Fatal(err)
		}

		fmt.Printf("\nresult :\n%s\n", result)
	*/

	result, err = mon.GetVMLatestMonMemoryInfoByParam(nsId, mcisId, vmId, "avg")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	result, err = mon.GetVMLatestMonDiskInfoByParam(nsId, mcisId, vmId, "avg")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	result, err = mon.GetVMLatestMonNetworkInfoByParam(nsId, mcisId, vmId, "avg")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	mon.Close()
}

// MonConfigApiTest - MonConfig 관련 API 호출 테스트
func MonConfigApiTest() {

	fmt.Print("\n\n============= MonConfigApiTest() =============\n")

	logger := logger.NewLogger()

	mon := api.NewMONManager()

	err := mon.SetConfigPath("./grpc_conf.yaml")
	if err != nil {
		logger.Fatal(err)
	}

	err = mon.Open()
	if err != nil {
		logger.Fatal(err)
	}

	reqMonitoring := &api.MonitoringReq{
		AgentInterval:      3,
		CollectorInterval:  15,
		SchedulingInterval: 10,
		MaxHostCount:       10,
		AgentTTL:           5,
	}

	result, err := mon.SetMonConfigByParam(reqMonitoring)
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	result, err = mon.GetMonConfig()
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	result, err = mon.ResetMonConfig()
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	mon.Close()
}

// InstallTelegrafApiTest - InstallTelegraf 관련 API 호출 테스트
func InstallTelegrafApiTest() {

	fmt.Print("\n\n============= InstallTelegrafApiTest() =============\n")

	logger := logger.NewLogger()

	mon := api.NewMONManager()

	err := mon.SetConfigPath("./grpc_conf.yaml")
	if err != nil {
		logger.Fatal(err)
	}

	err = mon.Open()
	if err != nil {
		logger.Fatal(err)
	}

	result, err := mon.InstallTelegrafByParam(nsId, mcisId, vmId, vmIp, vmUser, sshKey)
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("\nresult :\n%s\n", result)

	mon.Close()
}
