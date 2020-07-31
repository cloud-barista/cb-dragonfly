package agent

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	sshrun "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/cloud-barista/cb-dragonfly/pkg/util"
)

const (
	UBUNTU = "UBUNTU"
	CENTOS = "CENTOS"
)

func InstallTelegraf(nsId string, mcisId string, vmId string, publicIp string, userName string, sshKey string) error {
	sshInfo := sshrun.SSHInfo{
		ServerPort: publicIp + ":22",
		UserName:   userName,
		PrivateKey: []byte(sshKey),
	}

	// {사용자계정}/cb-dragonfly 폴더 생성
	createFolderCmd := fmt.Sprintf("mkdir $HOME/cb-dragonfly")
	if _, err := sshrun.SSHRun(sshInfo, createFolderCmd); err != nil {
		return errors.New(fmt.Sprintf("failed to make directory cb-dragonfly, error=%s", err))
	}

	// 리눅스 OS 환경 체크
	osType, err := sshrun.SSHRun(sshInfo, "hostnamectl | grep 'Operating System' | awk '{print $3}' | tr 'a-z' 'A-Z'")
	if err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to check linux OS environments, error=%s", err))
	}

	rootPath := os.Getenv("CBMON_ROOT")

	var sourceFile, targetFile, installCmd string
	if strings.Contains(osType, "CENTOS") {
		sourceFile = rootPath + "/file/pkg/centos/x64/telegraf-1.12.0~f09f2b5-0.x86_64.rpm"
		targetFile = fmt.Sprintf("$HOME/cb-dragonfly/cb-agent.rpm")
		installCmd = fmt.Sprintf("sudo rpm -ivh $HOME/cb-dragonfly/cb-agent.rpm")
	} else if strings.Contains(osType, "UBUNTU") {
		sourceFile = rootPath + "/file/pkg/ubuntu/x64/telegraf_1.12.0~f09f2b5-0_amd64.deb"
		targetFile = fmt.Sprintf("$HOME/cb-dragonfly/cb-agent.deb")
		installCmd = fmt.Sprintf("sudo dpkg -i $HOME/cb-dragonfly/cb-agent.deb")
	}

	// 에이전트 설치 패키지 다운로드
	if err := sshCopyWithTimeout(sshInfo, sourceFile, targetFile); err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to download agent package, error=%s", err))
	}

	// 패키지 설치 실행
	if _, err := sshrun.SSHRun(sshInfo, installCmd); err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to install agent package, error=%s", err))
	}

	sshrun.SSHRun(sshInfo, "sudo rm /etc/telegraf/telegraf.conf")

	// telegraf_conf 파일 복사
	telegrafConfSourceFile, err := createTelegrafConfigFile(nsId, mcisId, vmId)
	telegrafConfTargetFile := "$HOME/cb-dragonfly/telegraf.conf"
	if err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to create telegraf.conf, error=%s", err))
	}
	if err := sshrun.SSHCopy(sshInfo, telegrafConfSourceFile, telegrafConfTargetFile); err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to copy telegraf.conf, error=%s", err))
	}

	if _, err := sshrun.SSHRun(sshInfo, "sudo mv $HOME/cb-dragonfly/telegraf.conf /etc/telegraf/"); err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to move telegraf.conf, error=%s", err))
	}

	// 공통 서비스 활성화 및 실행
	if _, err := sshrun.SSHRun(sshInfo, "sudo systemctl enable telegraf && sudo systemctl restart telegraf"); err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to enable and start telegraf service, error=%s", err))
	}

	// telegraf UUId conf 파일 삭제
	err = os.Remove(telegrafConfSourceFile)
	if err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to remove temporary telegraf.conf file, error=%s", err))
	}

	// 에이전트 설치에 사용한 파일 폴더 채로 제거
	removeRpmCmd := fmt.Sprintf("sudo rm -rf $HOME/cb-dragonfly")
	if _, err := sshrun.SSHRun(sshInfo, removeRpmCmd); err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to remove cb-dragonfly directory, error=%s", err))
	}

	// 정상 설치 확인
	checkCmd := "telegraf --version"
	if result, err := util.RunCommand(publicIp, userName, sshKey, checkCmd); err != nil {
		cleanTelegrafInstall(sshInfo, osType)
		return errors.New(fmt.Sprintf("failed to run telegraf command, error=%s", err))
	} else {
		if strings.Contains(*result, "command not found") {
			cleanTelegrafInstall(sshInfo, osType)
			return errors.New(fmt.Sprintf("failed to run telegraf command, error=%s", err))
		}
		return nil
	}
}

func cleanTelegrafInstall(sshInfo sshrun.SSHInfo, osType string) {

	// Uninstall Telegraf
	var uninstallCmd string
	if strings.Contains(osType, "CENTOS") {
		uninstallCmd = fmt.Sprintf("sudo rpm -e telegraf")
	} else if strings.Contains(osType, "UBUNTU") {
		uninstallCmd = fmt.Sprintf("sudo dpkg -r telegraf")
	}
	sshrun.SSHRun(sshInfo, uninstallCmd)

	// Delete Install Files
	removeRpmCmd := fmt.Sprintf("sudo rm -rf $HOME/cb-dragonfly")
	sshrun.SSHRun(sshInfo, removeRpmCmd)
	removeDirCmd := fmt.Sprintf("sudo rm -rf /etc/telegraf/cb-dragonfly")
	sshrun.SSHRun(sshInfo, removeDirCmd)
}

func createTelegrafConfigFile(nsId string, mcisId string, vmId string) (string, error) {
	//collectorServer := fmt.Sprintf("udp://%s:%d", core.CoreConfig.Manager.Config.CollectManager.CollectorIP, core.CoreConfig.Manager.Config.CollectManager.CollectorPort)
	//influxDBServer := fmt.Sprintf("http://%s:8086", core.CoreConfig.Manager.Config.CollectManager.CollectorIP)
	collectorServer := "udp://127.0.0.1:8086"
	influxDBServer := "http://127.0.0.1:8086"

	rootPath := os.Getenv("CBMON_ROOT")
	filePath := rootPath + "/file/conf/telegraf.conf"

	read, err := ioutil.ReadFile(filePath)
	if err != nil {
		// ERROR 정보 출럭
		logrus.Error("failed to read telegraf.conf file.")
		return "", err
	}

	// 파일 내의 변수 값 설정 (hostId, collectorServer)
	strConf := string(read)
	strConf = strings.ReplaceAll(strConf, "{{ns_id}}", nsId)
	strConf = strings.ReplaceAll(strConf, "{{mcis_id}}", mcisId)
	strConf = strings.ReplaceAll(strConf, "{{vm_id}}", vmId)
	strConf = strings.ReplaceAll(strConf, "{{collector_server}}", collectorServer)
	strConf = strings.ReplaceAll(strConf, "{{influxdb_server}}", influxDBServer)

	// telegraf.conf 파일 생성
	telegrafFilePath := rootPath + "/file/conf/"
	createFileName := "telegraf-" + uuid.New().String() + ".conf"
	telegrafConfFile := telegrafFilePath + createFileName

	err = ioutil.WriteFile(telegrafConfFile, []byte(strConf), os.FileMode(777))
	if err != nil {
		logrus.Error("failed to create telegraf.conf file.")
		return "", err
	}

	return telegrafConfFile, err
}

func sshCopyWithTimeout(sshInfo sshrun.SSHInfo, sourceFile string, targetFile string) error {
	signer, err := ssh.ParsePrivateKey(sshInfo.PrivateKey)
	if err != nil {
		return err
	}
	clientConfig := ssh.ClientConfig{
		User: sshInfo.UserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client := scp.NewClientWithTimeout(sshInfo.ServerPort, &clientConfig, 600*time.Second)
	err = client.Connect()
	defer client.Close()
	if err != nil {
		return err
	}

	file, err := os.Open(sourceFile)
	defer file.Close()
	if err != nil {
		return err
	}

	return client.CopyFile(file, targetFile, "0755")
}
