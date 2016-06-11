package model

import (
	"bytes"
	"fmt"

	"github.com/bborbe/kubernetes_tools/config"
)

type Cluster struct {
	Host           string
	Region         string
	PublicIp       string
	LvmVolumeGroup string
	Network        string
	Gateway        string
	Dns            string
	Bridge         string
	Nodes          []*Node
}

type Node struct {
	Name        string
	Mac         string
	Ip          string
	VolumeName  string
	Etcd        bool
	Worker      bool
	Master      bool
	Nfsd        bool
	Storage     bool
	Cores       int
	Memory      int
	NfsSize     string
	StorageSize string
}

func NewCluster(cluster *config.Cluster) *Cluster {
	c := new(Cluster)

	c.Bridge = cluster.Bridge
	c.Region = cluster.Region
	c.Host = cluster.Host
	c.PublicIp = cluster.PublicIp
	c.LvmVolumeGroup = cluster.LvmVolumeGroup
	c.Network = cluster.Network
	c.Gateway = fmt.Sprintf("%s.1", cluster.Network)
	c.Dns = fmt.Sprintf("%s.1", cluster.Network)

	counter := 0
	for _, n := range cluster.Nodes {
		for i := 0; i < n.Amount; i++ {

			if n.Storage && n.Nfsd {
				panic("storage and nfsd at the same time is currently not supported")
			}

			name := generateNodeName(n, i)
			node := &Node{
				Name:        name,
				Ip:          fmt.Sprintf("%s.%d", cluster.Network, counter+10),
				Mac:         fmt.Sprintf("%s%02x", cluster.MacPrefix, counter+10),
				VolumeName:  fmt.Sprintf("%s%s", cluster.VolumePrefix, name),
				Etcd:        n.Etcd,
				Worker:      n.Worker,
				Master:      n.Master,
				Storage:     n.Storage,
				Nfsd:        n.Nfsd,
				Cores:       n.Cores,
				Memory:      n.Memory,
				NfsSize:     n.NfsSize,
				StorageSize: n.StorageSize,
			}
			c.Nodes = append(c.Nodes, node)
			counter++
		}
	}

	return c
}

func (c *Cluster) VolumeNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.VolumeName)
	}
	return result
}

func generateNodeName(node config.Node, number int) string {
	if node.Amount == 1 {
		return node.Name
	} else {
		return fmt.Sprintf("%s%d", node.Name, number)
	}
}

func (c *Cluster) NodeNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.Name)
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

func (n *Node) Roles() string {
	var roles []string
	if n.Etcd {
		roles = append(roles, "etcd")
	}
	if n.Worker {
		roles = append(roles, "worker")
	}
	if n.Master {
		roles = append(roles, "master")
	}
	if n.Storage {
		roles = append(roles, "storage")
	}
	content := bytes.NewBufferString("")
	for i, role := range roles {
		if i != 0 {
			content.WriteString(",")
		}
		content.WriteString("role=")
		content.WriteString(role)
	}
	return content.String()
}
