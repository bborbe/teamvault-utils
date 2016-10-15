package model_generator

import (
	"fmt"
	"github.com/bborbe/kubernetes_tools/config"
	"github.com/bborbe/kubernetes_tools/model"
)

const K8S_DEFAULT_VERSION = "v1.3.5"

func GenerateModel(cluster *config.Cluster) *model.Cluster {
	c := new(model.Cluster)

	c.UpdateRebootStrategy = cluster.UpdateRebootStrategy
	if len(c.UpdateRebootStrategy) == 0 {
		c.UpdateRebootStrategy = "etcd-lock"
	}
	c.Version = cluster.KubernetesVersion
	if len(c.Version) == 0 {
		c.Version = K8S_DEFAULT_VERSION
	}
	c.Region = cluster.Region
	c.LvmVolumeGroup = cluster.LvmVolumeGroup

	for _, host := range cluster.Hosts {
		//c.Host = cluster.Host
		//c.Bridge = cluster.Bridge
		//c.ApiServerPublicIp = cluster.ApiServerPublicIp
		//c.Network = cluster.Network
		//c.Gateway = valueOf(cluster.Gateway, fmt.Sprintf("%s.1", cluster.Network))
		//c.Dns = valueOf(cluster.Dns, fmt.Sprintf("%s.1", cluster.Network))

		counter := 0
		for _, n := range host.Nodes {
			for i := 0; i < n.Amount; i++ {

				if n.Storage && n.Nfsd {
					panic("storage and nfsd at the same time is currently not supported")
				}

				//name := generateNodeName(n, i)
				node := &model.Node{
					//Name:        name,
					Ip:  fmt.Sprintf("%s.%d", host.KubernetesNetwork, counter+10),
					Mac: fmt.Sprintf("%s%02x", host.KubernetesNetwork, counter+10),
					//VolumeName:  fmt.Sprintf("%s%s", cluster.VolumePrefix, name),
					//VmName:      fmt.Sprintf("%s%s", cluster.VmPrefix, name),
					Etcd:        n.Etcd,
					Worker:      n.Worker,
					Master:      n.Master,
					Storage:     n.Storage,
					Nfsd:        n.Nfsd,
					Cores:       n.Cores,
					Memory:      n.Memory,
					NfsSize:     n.NfsSize,
					StorageSize: n.StorageSize,
					RootSize:    valueOfSize(n.RootSize, "10G"),
					DockerSize:  valueOfSize(n.DockerSize, "10G"),
					KubeletSize: valueOfSize(n.KubeletSize, "10G"),
				}
				c.Nodes = append(c.Nodes, node)
				counter++
			}
		}
	}

	return c
}

func valueOfSize(size model.Size, defaultSize model.Size) model.Size {
	if len(size) == 0 {
		return defaultSize
	}
	return size
}

//func generateNodeName(node config.Node, number int) string {
//	if node.Amount == 1 {
//		return node.Name
//	} else {
//		return fmt.Sprintf("%s%d", node.Name, number)
//	}
//}

func valueOf(value string, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
