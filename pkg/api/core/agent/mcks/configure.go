package mcks

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/cloud-barista/cb-dragonfly/pkg/api/core/agent/common"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/types"
	"github.com/cloud-barista/cb-dragonfly/pkg/util"
	"io"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kubeserialize "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	k8sRestClient "k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"net/http"
	"os"
	"strings"
)

func CreateTelegrafConfigConfigmap(info common.AgentInstallInfo, yamlData unstructured.Unstructured) (corev1.ConfigMap, error) {
	mechanism := fmt.Sprintf(strings.ToLower(config.GetInstance().Monitoring.DefaultPolicy))
	if strings.EqualFold(mechanism, common.PULL_MECHANISM) {
		return corev1.ConfigMap{}, errors.New("pull monitoring for mcks is not supported")
	}
	rootPath := os.Getenv("CBMON_ROOT")
	filePath := rootPath + "/file/conf/mcks/telegraf.conf"

	read, err := ioutil.ReadFile(filePath)
	if err != nil {
		// ERROR 정보 출럭
		util.GetLogger().Error("failed to read telegraf.conf file.")
		return corev1.ConfigMap{}, err
	}

	// 파일 내의 변수 값 설정 (hostId, collectorServer)
	strConf := string(read)

	// 파일 MCKS 에이전트 변수 값 설정
	strConf = strings.ReplaceAll(strConf, "{{topic}}", fmt.Sprintf("%s_mcks_%s", info.NsId, info.McksID))
	strConf = strings.ReplaceAll(strConf, "{{ns_id}}", info.NsId)
	strConf = strings.ReplaceAll(strConf, "{{mcks_id}}", info.McksID)
	strConf = strings.ReplaceAll(strConf, "{{mechanism}}", mechanism)

	var kafkaPort int
	if config.GetInstance().GetMonConfig().DeployType == types.Helm {
		kafkaPort = config.GetInstance().Kafka.HelmPort
	} else {
		kafkaPort = types.KafkaDefaultPort
	}
	kafkaAddr := fmt.Sprintf("%s:%d", config.GetInstance().Kafka.EndpointUrl, kafkaPort)
	strConf = strings.ReplaceAll(strConf, "{{broker_server}}", kafkaAddr)

	// 컨피그맵 기본 데이터 설정
	if err != nil {
		return corev1.ConfigMap{}, errors.New(fmt.Sprintf("no such file '%s', error=%s", filePath, err))
	}

	agentConfInfo := corev1.ConfigMap{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(yamlData.Object, &agentConfInfo); err != nil {
		return corev1.ConfigMap{}, err
	}

	agentConfInfo.Data = map[string]string{
		"telegraf.conf": strConf,
	}

	return agentConfInfo, nil
}

func ConfigAgentDaemonSetHostAlias(yamlData unstructured.Unstructured, labels map[string]string) (map[string]interface{}, error) {
	agentDaemonSetInfo := appsv1.DaemonSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(yamlData.Object, &agentDaemonSetInfo); err != nil {
		return yamlData.Object, err
	}

	agentDaemonSetInfo.Spec.Template.Spec.HostAliases = []corev1.HostAlias{
		{
			IP:        config.GetInstance().Dragonfly.DragonflyIP,
			Hostnames: []string{config.GetInstance().Kafka.EndpointUrl},
		},
	}
	// 라벨 설정
	agentDaemonSetInfo.Spec.Selector.MatchLabels = labels
	agentDaemonSetInfo.Spec.Template.ObjectMeta.SetLabels(labels)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&agentDaemonSetInfo)
	if err != nil {
		return yamlData.Object, err
	}
	return obj, nil
}

func InstallAgent(info common.AgentInstallInfo) (int, error) {
	serverCA, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(info.ServerCA))
	clientCert, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(info.ClientCA))
	clientKey, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(info.ClientKey))
	clientToken, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(info.ClientToken))

	kubeconfig := &k8sRestClient.Config{
		Host: info.APIServerURL,
		TLSClientConfig: k8sRestClient.TLSClientConfig{
			CAData: serverCA,
		},
	}

	if len(info.ClientToken) == 0 {
		kubeconfig.TLSClientConfig.CertData = clientCert
		kubeconfig.TLSClientConfig.KeyData = clientKey
	} else {
		kubeconfig.BearerToken = string(clientToken)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create kubeclient, error=%s", err))
	}

	namespaceInfo, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), common.AGENT_NAMESPACE, metav1.GetOptions{})
	if err != nil {
		// 네임스페이스가 없을 경우 생성
		if apierrors.IsNotFound(err) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: common.AGENT_NAMESPACE,
				},
			}
			namespaceInfo, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
			if err != nil {
				return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create namespace for deploying agent, error=%s", err))
			}
		}
	}

	rootPath := os.Getenv("CBMON_ROOT")
	commonDir := rootPath + "/file/agent/mcks"

	dynamicClient, err := dynamic.NewForConfig(kubeconfig)
	if err != nil {
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create dynamic client, error=%s", err))
	}

	fileNameList, err := common.GetAllFilesinPath(commonDir)
	if err != nil {
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("no files exist in %s, error=%s", commonDir, err))
	}

	for _, filename := range fileNameList {
		filePath := fmt.Sprintf("%s/%s", commonDir, filename)

		file, err := os.Open(filePath)
		if err != nil {
			return http.StatusInternalServerError, errors.New(fmt.Sprintf("cannot open yaml file %s, err=%s", filename, err))
		}

		decoder := kubeyaml.NewYAMLOrJSONDecoder(file, 4096)
		for {
			ext := runtime.RawExtension{}
			if err = decoder.Decode(&ext); err != nil {
				if err == io.EOF {
					break
				}
			}

			u := unstructured.Unstructured{}
			_, gvr, err := kubeserialize.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(ext.Raw, nil, &u)
			agentLabel := map[string]string{
				"controller": "cb-dragonfly",
				"app":        "telegraf",
			}

			// 전체 리소스 라벨 생성
			u.SetLabels(agentLabel)

			// 컨피그맵일 경우 데이터 설정 및 생성
			if strings.EqualFold(u.GetKind(), "configmap") {
				configmap, err := CreateTelegrafConfigConfigmap(info, u)
				if err != nil {
					return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create agent configuration configmap info, err=%s", err))
				}

				if _, err = kubeClient.CoreV1().ConfigMaps(namespaceInfo.Name).Create(context.TODO(), &configmap, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}}); err != nil {
					return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create agent configuration configmap, err=%s", err))
				}
				if _, err = kubeClient.CoreV1().ConfigMaps(namespaceInfo.Name).Create(context.TODO(), &configmap, metav1.CreateOptions{}); err != nil {
					common.CleanAgentInstall(info, nil, nil, kubeClient)
					return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create configmap, err=%s", err))
				}
				continue
			}

			if strings.EqualFold(u.GetKind(), "daemonset") {
				if u.Object, err = ConfigAgentDaemonSetHostAlias(u, agentLabel); err != nil {
					common.CleanAgentInstall(info, nil, nil, kubeClient)
					return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to config hostalias for agent daemonset, err=%s", err))
				}
			}
			mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(kubeClient.DiscoveryClient))
			mapping, err := mapper.RESTMapping(gvr.GroupKind(), gvr.Version)

			// 클러스터 Scope 리소스인지 확인
			var dynamicResource dynamic.ResourceInterface
			if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
				// Namespaced resource
				dynamicResource = dynamicClient.Resource(mapping.Resource).Namespace(namespaceInfo.Name)
			} else {
				// Cluster-wide resource
				dynamicResource = dynamicClient.Resource(mapping.Resource)
			}

			if _, err := dynamicResource.List(context.TODO(), metav1.ListOptions{}); err != nil {
				return http.StatusInternalServerError, errors.New(fmt.Sprintf("error with %s, '%s', err=%s", u.GetKind(), u.GetName(), err))
			}
			// 그 외의 데이터 생성
			if resource, _ := dynamicResource.Get(context.TODO(), u.GetName(), metav1.GetOptions{}); resource != nil {
				common.CleanAgentInstall(info, nil, nil, kubeClient)
				return http.StatusInternalServerError, errors.New(fmt.Sprintf("already exist resource %s, '%s'", u.GetKind(), u.GetName()))
			}
			_, err = dynamicResource.Create(context.TODO(), &u, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}})
			if err != nil {
				return http.StatusInternalServerError, errors.New(fmt.Sprintf("error with %s, '%s', err=%s", u.GetKind(), u.GetName(), err))
			}
			if _, err = dynamicResource.Create(context.TODO(), &u, metav1.CreateOptions{}); err != nil {
				common.CleanAgentInstall(info, nil, nil, kubeClient)
				return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create %s, '%s', err=%s", u.GetKind(), u.GetName(), err))
			}
		}
	}

	// 메타데이터 저장
	if _, _, err = common.PutAgent(info); err != nil {
		common.CleanAgentInstall(info, nil, nil, kubeClient)
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to put metadata to cb-store, error=%s", err))
	}
	return http.StatusOK, nil
}

func UninstallAgent(info common.AgentInstallInfo) (int, error) {
	serverCA, _ := base64.StdEncoding.DecodeString(info.ServerCA)
	clientCert, _ := base64.StdEncoding.DecodeString(info.ClientCA)
	clientKey, _ := base64.StdEncoding.DecodeString(info.ClientKey)
	clientToken, _ := base64.StdEncoding.DecodeString(info.ClientToken)

	kubeconfig := k8sRestClient.Config{
		Host: info.APIServerURL,
	}

	if len(info.ClientToken) == 0 {
		kubeconfig.TLSClientConfig.CAData = serverCA
		kubeconfig.TLSClientConfig.CertData = clientCert
		kubeconfig.TLSClientConfig.KeyData = clientKey
	} else {
		kubeconfig.BearerToken = string(clientToken)
	}

	kubeClient, err := kubernetes.NewForConfig(&kubeconfig)
	if err != nil {
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to create kubeclient, error=%s", err))
	}

	// 에이전트 관련 리소스 삭제
	common.CleanAgentInstall(info, nil, nil, kubeClient)

	// 메타데이터 삭제
	if err = common.DeleteAgent(info); err != nil {
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("failed to delete metadata, error=%s", err))
	}
	return http.StatusOK, nil
}
