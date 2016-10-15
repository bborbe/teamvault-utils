package model

import (
	"bytes"
	"strings"
)

type UpdateRebootStrategy string
type KubernetesVersion string
type Region string
type VmPrefix string
type VolumePrefix string
type LvmVolumeGroup string
type Size string
type Device string
type Gateway string
type Network string
type HostName string

type Cluster struct {
	Version              KubernetesVersion
	Host                 string
	Region               Region
	ApiServerPublicIp    string
	LvmVolumeGroup       LvmVolumeGroup
	Network              string
	Gateway              string
	Dns                  string
	Bridge               string
	Nodes                []*Node
	UpdateRebootStrategy UpdateRebootStrategy
}

type Node struct {
	Name        string
	Mac         string
	Ip          string
	VolumeName  string
	VmName      string
	Etcd        bool
	Worker      bool
	Master      bool
	Nfsd        bool
	Storage     bool
	Cores       int
	Memory      int
	NfsSize     Size
	StorageSize Size
	RootSize    Size
	DockerSize  Size
	KubeletSize Size
}

func (c *Cluster) VolumeNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.VolumeName)
	}
	return result
}

func (c *Cluster) NodeNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.Name)
	}
	return result
}

func (c *Cluster) VmNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.VmName)
	}
	return result
}

func (c *Cluster) MasterNodes() []*Node {
	var result []*Node
	for _, node := range c.Nodes {
		if node.Master {
			result = append(result, node)
		}
	}
	return result
}

func (c *Cluster) NotMasterNodes() []*Node {
	var result []*Node
	for _, node := range c.Nodes {
		if !node.Master {
			result = append(result, node)
		}
	}
	return result
}

func (c *Cluster) StorageNodes() []*Node {
	var result []*Node
	for _, node := range c.Nodes {
		if node.Storage {
			result = append(result, node)
		}
	}
	return result
}

func (c *Cluster) NfsdNodes() []*Node {
	var result []*Node
	for _, node := range c.Nodes {
		if node.Nfsd {
			result = append(result, node)
		}
	}
	return result
}

func (c *Cluster) EtcdEndpoints() string {
	first := true
	content := bytes.NewBufferString("")
	for _, node := range c.Nodes {
		if node.Etcd {
			if first {
				first = false
			} else {
				content.WriteString(",")
			}
			content.WriteString("http://")
			content.WriteString(node.Ip)
			content.WriteString(":2379")
		}
	}
	return content.String()
}

func (c *Cluster) InitialCluster() string {
	first := true
	content := bytes.NewBufferString("")
	for _, node := range c.Nodes {
		if node.Etcd {
			if first {
				first = false
			} else {
				content.WriteString(",")
			}
			content.WriteString(node.Name)
			content.WriteString("=http://")
			content.WriteString(node.Ip)
			content.WriteString(":2380")
		}
	}
	return content.String()
}

func (c *Cluster) ApiServers() string {
	first := true
	content := bytes.NewBufferString("")
	for _, node := range c.Nodes {
		if node.Master {
			if first {
				first = false
			} else {
				content.WriteString(",")
			}
			content.WriteString("https://")
			content.WriteString(node.Ip)
		}
	}
	return content.String()
}

func (n *Node) Labels() string {
	var labels []string
	if n.Etcd {
		labels = append(labels, "etcd=true")
	}
	if n.Storage {
		labels = append(labels, "storage=true")
	}
	if n.Nfsd {
		labels = append(labels, "nfsd=true")
	}
	if n.Worker {
		labels = append(labels, "worker=true")
	}
	if n.Master {
		labels = append(labels, "master=true")
	}
	return strings.Join(labels, ",")
}
