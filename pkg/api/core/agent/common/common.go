package common

import (
	"context"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	sshrun "github.com/cloud-barista/cb-spider/cloud-control-manager/vm-ssh"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	UBUNTU                   = "UBUNTU"
	CENTOS                   = "CENTOS"
	AGENT_NAMESPACE          = "cb-dragonfly"
	AGENT_CLUSTERROLE        = "cb-dragonfly-agent-clusterrole"
	AGENT_CLUSTERROLEBINDING = "cb-dragonfly-agent-clusterrolebinding"
	PULL_MECHANISM           = "pull"
)

type AgentInstallInfo struct {
	NsId         string
	McisId       string
	VmId         string
	PublicIp     string
	UserName     string
	SshKey       string
	CspType      string
	Port         string
	ServiceType  string
	Mck8sId      string
	APIServerURL string
	ServerCA     string
	ClientCA     string
	ClientKey    string
	ClientToken  string
}

func CleanAgentInstall(info AgentInstallInfo, sshInfo *sshrun.SSHInfo, osType *string, kubeClient *kubernetes.Clientset) {
	if util.CheckMCK8SType(info.ServiceType) {
		_ = kubeClient.RbacV1().ClusterRoleBindings().Delete(context.TODO(), AGENT_CLUSTERROLEBINDING, metav1.DeleteOptions{})
		_ = kubeClient.RbacV1().ClusterRoles().Delete(context.TODO(), AGENT_CLUSTERROLE, metav1.DeleteOptions{})
		_ = kubeClient.CoreV1().Namespaces().Delete(context.TODO(), config.GetInstance().Agent.Namespace, metav1.DeleteOptions{})
		return
	}
	// Uninstall Telegraf
	var uninstallCmd string
	if strings.Contains(*osType, CENTOS) {
		uninstallCmd = fmt.Sprintf("sudo rpm -e telegraf")
	} else if strings.Contains(*osType, UBUNTU) {
		uninstallCmd = fmt.Sprintf("sudo dpkg -r telegraf")
	}
	sshrun.SSHRun(*sshInfo, uninstallCmd)

	// Delete Install Files
	removeRpmCmd := fmt.Sprintf("sudo rm -rf $HOME/cb-dragonfly")
	sshrun.SSHRun(*sshInfo, removeRpmCmd)
	removeDirCmd := fmt.Sprintf("sudo rm -rf /etc/telegraf/telegraf.conf")
	sshrun.SSHRun(*sshInfo, removeDirCmd)
}

func GetPackageName(path string) (string, error) {
	file, err := ioutil.ReadDir(path)
	var filename string
	for _, data := range file {
		filename = data.Name()
	}
	return filename, err
}

func GetAllFilesinPath(path string) ([]string, error) {
	file, err := ioutil.ReadDir(path)
	fileNameList := []string{}
	for _, data := range file {
		fileNameList = append(fileNameList, data.Name())
	}
	return fileNameList, err
}
