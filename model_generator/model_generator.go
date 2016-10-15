package model_generator

import (
	"github.com/bborbe/kubernetes_tools/config"
	"github.com/bborbe/kubernetes_tools/model"
)

const K8S_DEFAULT_VERSION = "v1.3.5"

func GenerateModel(config *config.Cluster) (*model.Cluster, error) {
	cluster := new(model.Cluster)

	dnsIp, err := model.IpByString("8.8.8.8")
	if err != nil {
		return nil, err
	}
	dns := model.Dns(*dnsIp)

	cluster.UpdateRebootStrategy = config.UpdateRebootStrategy
	if len(cluster.UpdateRebootStrategy) == 0 {
		cluster.UpdateRebootStrategy = "etcd-lock"
	}
	cluster.Version = config.KubernetesVersion
	if len(cluster.Version) == 0 {
		cluster.Version = K8S_DEFAULT_VERSION
	}
	cluster.Region = config.Region
	for _, configHost := range config.Hosts {
		//	cluster.LvmVolumeGroup = config.LvmVolumeGroup
		//c.Bridge = cluster.Bridge
		//c.ApiServerPublicIp = cluster.ApiServerPublicIp
		//c.Network = cluster.Network
		//c.Gateway = valueOf(cluster.Gateway, fmt.Sprintf("%s.1", cluster.Network))
		//c.Dns = valueOf(cluster.Dns, fmt.Sprintf("%s.1", cluster.Network))

		host := model.Host{}
		host.Name = configHost.Name
		host.LvmVolumeGroup = configHost.LvmVolumeGroup
		host.VolumePrefix = configHost.VolumePrefix
		host.VmPrefix = configHost.VmPrefix

		gatewayIp := configHost.KubernetesNetwork.Ip
		gatewayIp.Set(3, 1)
		gateway := model.Gateway(gatewayIp)
		counter := 0
		for _, configNode := range configHost.Nodes {
			for i := 0; i < configNode.Amount; i++ {

				if configNode.Storage && configNode.Nfsd {
					panic("storage and nfsd at the same time is currently not supported")
				}

				address := configHost.KubernetesNetwork
				address.Ip.Set(3, byte(counter+10))
				mac, err := address.Ip.Mac()
				if err != nil {
					return nil, err
				}
				//name := generateNodeName(n, i)
				node := model.Node{
					KuberntesNetwork: &model.Network{
						Device:  configHost.KubernetesDevice,
						Address: address,
						Mac:     *mac,
						Gateway: gateway,
						Dns:     dns,
					},
					//Name:        name,
					//Ip:  fmt.Sprintf("%s.%d", configHost.KubernetesNetwork, counter + 10),
					//Mac: fmt.Sprintf("%s%02x", configHost.KubernetesNetwork, counter + 10),
					//VolumeName:  fmt.Sprintf("%s%s", cluster.VolumePrefix, name),
					//VmName:      fmt.Sprintf("%s%s", cluster.VmPrefix, name),
					Etcd:        configNode.Etcd,
					Worker:      configNode.Worker,
					Master:      configNode.Master,
					Storage:     configNode.Storage,
					Nfsd:        configNode.Nfsd,
					Cores:       configNode.Cores,
					Memory:      configNode.Memory,
					NfsSize:     configNode.NfsSize,
					StorageSize: configNode.StorageSize,
					RootSize:    valueOfSize(configNode.RootSize, "10G"),
					DockerSize:  valueOfSize(configNode.DockerSize, "10G"),
					KubeletSize: valueOfSize(configNode.KubeletSize, "10G"),
				}
				host.Nodes = append(host.Nodes, node)
				counter++
			}
		}
		cluster.Hosts = append(cluster.Hosts, host)
	}

	return cluster, nil
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
