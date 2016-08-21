package model

import (
	"testing"

	. "github.com/bborbe/assert"

	"github.com/bborbe/kubernetes_tools/config"
)

func TestNetwork(t *testing.T) {
	cluster := &config.Cluster{
		Network: "192.168.20",
	}
	object := NewCluster(cluster)

	if err := AssertThat(object.Network, Is("192.168.20")); err != nil {
		t.Fatal(err)
	}
}

func TestDns(t *testing.T) {
	cluster := &config.Cluster{
		Network: "192.168.20",
		Dns:     "192.168.20.123",
	}
	object := NewCluster(cluster)

	if err := AssertThat(object.Dns, Is("192.168.20.123")); err != nil {
		t.Fatal(err)
	}
}

func TestDnsDefault(t *testing.T) {
	cluster := &config.Cluster{
		Network: "192.168.20",
		Dns:     "",
	}
	object := NewCluster(cluster)

	if err := AssertThat(object.Dns, Is("192.168.20.1")); err != nil {
		t.Fatal(err)
	}
}

func TestGateway(t *testing.T) {
	cluster := &config.Cluster{
		Network: "192.168.20",
		Gateway: "192.168.20.123",
	}
	object := NewCluster(cluster)

	if err := AssertThat(object.Gateway, Is("192.168.20.123")); err != nil {
		t.Fatal(err)
	}
}

func TestGatewayDefault(t *testing.T) {
	cluster := &config.Cluster{
		Network: "192.168.20",
		Gateway: "",
	}
	object := NewCluster(cluster)

	if err := AssertThat(object.Gateway, Is("192.168.20.1")); err != nil {
		t.Fatal(err)
	}
}
