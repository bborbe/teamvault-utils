package generator

import (
	"fmt"

	"github.com/bborbe/kubernetes_tools/cluster/config"
	"github.com/bborbe/kubernetes_tools/cluster/model"
	"github.com/golang/glog"
)

const K8S_DEFAULT_VERSION = "v1.5.6"

func Generate(configCluster *config.Cluster) (*model.Cluster, error) {
	cluster := new(model.Cluster)

	cluster.UpdateRebootStrategy = configCluster.UpdateRebootStrategy
	if len(cluster.UpdateRebootStrategy) == 0 {
		cluster.UpdateRebootStrategy = "etcd-lock"
	}
	cluster.Version = configCluster.KubernetesVersion
	if len(cluster.Version) == 0 {
		cluster.Version = K8S_DEFAULT_VERSION
	}
	cluster.Region = configCluster.Region
	for _, configHost := range configCluster.Hosts {
		host, err := createHost(configHost)
		if err != nil {
			return nil, fmt.Errorf("create host failed: %v", err)
		}
		cluster.Hosts = append(cluster.Hosts, *host)
	}
	return cluster, nil
}

func createHost(configHost config.Host) (*model.Host, error) {
	host := model.Host{}
	host.Name = configHost.Name
	host.LvmVolumeGroup = configHost.LvmVolumeGroup
	host.VolumePrefix = configHost.VolumePrefix
	host.VmPrefix = configHost.VmPrefix

	counter := 0
	for _, configNode := range configHost.Nodes {

		if configNode.Storage && configNode.Nfsd {
			return nil, fmt.Errorf("storage and nfsd at the same time is currently not supported")
		}

		for i := 0; i < max(configNode.Amount, 1); i++ {

			node := model.Node{
				KubernetesNetwork: &model.Network{
					Number: 3,
					Device: configHost.KubernetesDevice,
				},
				Name:                generateNodeName(configNode, i),
				VolumeName:          generateVolumeName(configHost, configNode, i),
				VmName:              generateVmName(configHost, configNode, i),
				Etcd:                configNode.Etcd,
				Worker:              configNode.Worker,
				Master:              configNode.Master,
				Storage:             configNode.Storage,
				Nfsd:                configNode.Nfsd,
				Cores:               configNode.Cores,
				Memory:              configNode.Memory,
				NfsSize:             configNode.NfsSize,
				StorageSize:         configNode.StorageSize,
				RootSize:            valueOfSize(configNode.RootSize, "10G"),
				DockerSize:          valueOfSize(configNode.DockerSize, "10G"),
				KubeletSize:         valueOfSize(configNode.KubeletSize, "10G"),
				ApiServerPort:       valueOfInt(configNode.ApiServerPort, 443),
				IptablesFilterRules: configNode.IptablesFilterRules,
				IptablesNatRules:    configNode.IptablesNatRules,
			}

			dns, err := createDns(configHost, configNode)
			if err == nil {
				node.KubernetesNetwork.Dns = *dns
			}

			gateway, err := createGateway(configHost, configNode)
			if err == nil {
				node.KubernetesNetwork.Gateway = *gateway
			}

			address, err := createIp(configHost, configNode, counter)
			if err == nil {
				node.KubernetesNetwork.Address = *address
			}

			mac, err := createMac(configHost, configNode, counter)
			if err == nil {
				node.KubernetesNetwork.Mac = *mac
			}

			host.Nodes = append(host.Nodes, node)
			counter++
		}
	}
	return &host, nil
}

func createMac(configHost config.Host, node config.Node, counter int) (*model.Mac, error) {
	if len(node.Mac) > 0 {
		return model.MacByString(node.Mac)
	}
	address, err := createIp(configHost, node, counter)
	if err != nil {
		return nil, err
	}
	mac, err := address.Ip.Mac()
	if err != nil {
		glog.V(2).Infof("get mac failed: %v", err)
		return nil, fmt.Errorf("get mac failed: %v", err)
	}
	glog.V(2).Infof("kubernetes mac: %s", mac.String())
	return mac, nil
}

func createIp(configHost config.Host, node config.Node, counter int) (*model.Address, error) {
	if len(node.Address) > 0 {
		return model.ParseAddress(node.Address)
	}
	kubernetesNetwork, err := createNetwork(configHost)
	if err != nil {
		return nil, err
	}
	address := *kubernetesNetwork
	address.Ip.Set(15, byte(counter+10))
	glog.V(2).Infof("kubernetes address: %s", address.String())
	return &address, nil
}

func createDns(host config.Host, node config.Node) (*model.Dns, error) {
	if len(node.Dns) > 0 {
		ip, err := model.IpByString(node.Dns)
		if err != nil {
			return nil, err
		}
		dns := model.Dns(*ip)
		return &dns, nil
	}
	dnsIp, err := model.IpByString(host.KubernetesDns)
	if err != nil {
		glog.V(2).Infof("parse dns failed: %v", err)
		return nil, fmt.Errorf("parse dns failed: %v", err)
	}
	dns := model.Dns(*dnsIp)
	return &dns, nil
}

func createGateway(configHost config.Host, node config.Node) (*model.Gateway, error) {
	if len(node.Gateway) > 0 {
		ip, err := model.IpByString(node.Gateway)
		if err != nil {
			return nil, err
		}
		gateway := model.Gateway(*ip)
		return &gateway, nil
	}
	kubernetesNetwork, err := createNetwork(configHost)
	if err != nil {
		return nil, err
	}
	gatewayIp := kubernetesNetwork.Ip
	gatewayIp.Set(15, 1)
	gateway := model.Gateway(gatewayIp)
	glog.V(2).Infof("gateway %s", gateway.String())
	return &gateway, nil
}

func createNetwork(configHost config.Host) (*model.Address, error) {
	kubernetesNetwork, err := model.ParseAddress(configHost.KubernetesNetwork)
	if err != nil {
		glog.V(2).Infof("parse kubernetes network failed: %v", err)
		return nil, err
	}
	glog.V(2).Infof("kubernetes network: %s", kubernetesNetwork.String())
	glog.V(2).Infof("kubernetes ip: %s", kubernetesNetwork.Ip.String())
	glog.V(2).Infof("kubernetes mask: %s", kubernetesNetwork.Mask.String())
	return kubernetesNetwork, nil
}

func valueOfSize(size model.Size, defaultSize model.Size) model.Size {
	if len(size) == 0 {
		return defaultSize
	}
	return size
}

func valueOfInt(value int, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func generateNodeName(node config.Node, number int) model.NodeName {
	if node.Amount > 1 {
		return model.NodeName(fmt.Sprintf("%s%d", node.Name, number))
	}
	return model.NodeName(node.Name)
}

func generateVolumeName(host config.Host, node config.Node, number int) model.VolumeName {
	return model.VolumeName(fmt.Sprintf("%s%s", host.VolumePrefix, generateNodeName(node, number)))
}

func generateVmName(host config.Host, node config.Node, number int) model.VmName {
	return model.VmName(fmt.Sprintf("%s%s", host.VmPrefix, generateNodeName(node, number)))
}

func max(a int, bs ...int) int {
	result := a
	for _, b := range bs {
		if b > result {
			result = b
		}
	}
	return result
}
