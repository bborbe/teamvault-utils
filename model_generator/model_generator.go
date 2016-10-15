package model_generator

import (
	"fmt"
	"github.com/bborbe/kubernetes_tools/config"
	"github.com/bborbe/kubernetes_tools/model"
	"github.com/golang/glog"
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
		host := model.Host{}
		host.Name = configHost.Name
		host.LvmVolumeGroup = configHost.LvmVolumeGroup
		host.VolumePrefix = configHost.VolumePrefix
		host.VmPrefix = configHost.VmPrefix

		kubernetesNetwork, err := model.ParseAddress(configHost.KubernetesNetwork)
		if err != nil {
			return nil, err
		}
		glog.V(2).Infof("kubernetes network: %s", kubernetesNetwork.String())
		glog.V(2).Infof("kubernetes ip: %s", kubernetesNetwork.Ip.String())
		glog.V(2).Infof("kubernetes mask: %s", kubernetesNetwork.Mask.String())

		gatewayIp := kubernetesNetwork.Ip
		gatewayIp.Set(15, 1)
		gateway := model.Gateway(gatewayIp)
		glog.V(2).Infof("gateway %s", gateway.String())
		counter := 0
		for _, configNode := range configHost.Nodes {
			for i := 0; i < configNode.Amount; i++ {

				if configNode.Storage && configNode.Nfsd {
					panic("storage and nfsd at the same time is currently not supported")
				}

				address := *kubernetesNetwork
				address.Ip.Set(15, byte(counter+10))
				glog.V(2).Infof("kubernetes address: %s", address.String())
				mac, err := address.Ip.Mac()
				if err != nil {
					return nil, err
				}
				glog.V(2).Infof("kubernetes mac: %s", mac.String())
				node := model.Node{
					KubernetesNetwork: &model.Network{
						Number:  3,
						Device:  configHost.KubernetesDevice,
						Address: address,
						Mac:     *mac,
						Gateway: gateway,
						Dns:     dns,
					},
					Name:       generateNodeName(configNode, i),
					VolumeName: generateVolumeName(configHost, configNode, i),
					VmName:     generateVmName(configHost, configNode, i),
					//Ip:  fmt.Sprintf("%s.%d", configHost.KubernetesNetwork, counter + 10),
					//Mac: fmt.Sprintf("%s%02x", configHost.KubernetesNetwork, counter + 10),
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

func generateNodeName(node config.Node, number int) model.NodeName {
	if node.Amount == 1 {
		return model.NodeName(node.Name)
	} else {
		return model.NodeName(fmt.Sprintf("%s%d", node.Name, number))
	}
}

func generateVolumeName(host config.Host, node config.Node, number int) model.VolumeName {
	return model.VolumeName(fmt.Sprintf("%s%s", host.VolumePrefix, generateNodeName(node, number)))
}

func generateVmName(host config.Host, node config.Node, number int) model.VmName {
	return model.VmName(fmt.Sprintf("%s%s", host.VmPrefix, generateNodeName(node, number)))
}
