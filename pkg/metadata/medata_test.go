package metadata

import (
	"fmt"
	"testing"
)

func TestAddAgent(t *testing.T) {
	var agentListManager AgentListManager

	testAgent01 := []string{"test_ns01", "test_mcis01", "test_vm01", "aws", "192.168.1.1"}
	err := AgentInstallationMetadata(testAgent01[0], testAgent01[1], testAgent01[2], testAgent01[3], testAgent01[4])
	if err != nil {
		t.Error(err)
	}

	agentInfo, err := agentListManager.GetAgentInfo(fmt.Sprintf("%s/%s/%s/%s", testAgent01[0], testAgent01[1], testAgent01[2], testAgent01[3]))
	if err != nil {
		t.Error(err)
	}
	t.Log(agentInfo)
	agentList, err := agentListManager.GetAgentList()
	if err != nil {
		t.Error(err)
	}
	t.Log(agentList)

	testAgent02 := []string{"test_ns02", "test_mcis02", "test_vm02", "openstack", "192.168.1.2"}
	err = AgentInstallationMetadata(testAgent02[0], testAgent02[1], testAgent02[2], testAgent02[3], testAgent02[4])
	if err != nil {
		t.Error(err)
	}

	agentInfo, err = agentListManager.GetAgentInfo(fmt.Sprintf("%s/%s/%s/%s", testAgent02[0], testAgent02[1], testAgent02[2], testAgent02[3]))
	if err != nil {
		t.Error(err)
	}
	t.Log(agentInfo)
	agentList, err = agentListManager.GetAgentList()
	if err != nil {
		t.Error(err)
	}
	t.Log(agentList)
}

func TestDeleteAgent(t *testing.T) {
	var agentListManager AgentListManager

	testAgent01 := []string{"test_ns01", "test_mcis01", "test_vm01", "aws", "192.168.1.1"}
	err := AgentDeletionMetadata(testAgent01[0], testAgent01[1], testAgent01[2], testAgent01[3], testAgent01[4])
	if err != nil {
		t.Error(err)
	}

	agentList, err := agentListManager.GetAgentList()
	if err != nil {
		t.Error(err)
	}
	t.Log(agentList)

	testAgent02 := []string{"test_ns02", "test_mcis02", "test_vm02", "openstack", "192.168.1.2"}
	err = AgentDeletionMetadata(testAgent02[0], testAgent02[1], testAgent02[2], testAgent02[3], testAgent02[4])
	if err != nil {
		t.Error(err)
	}

	agentList, err = agentListManager.GetAgentList()
	if err != nil {
		t.Error(err)
	}
	t.Log(agentList)
}
