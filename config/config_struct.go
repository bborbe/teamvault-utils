package config

import (
	"github.com/bborbe/kubernetes_tools/model"
)

type Cluster struct {
	UpdateRebootStrategy model.UpdateRebootStrategy `json:"update-reboot-strategy"`
	KubernetesVersion    model.KubernetesVersion    `json:"kuberntes-version"`
	Region               model.Region               `json:"region"`
	Hosts                []Host                     `json:"hosts"`
	Features             Features                   `json:"features"`
}

type Features struct {
	Kvm      bool `json:"kvm"`
	Iptables bool `json:"iptables"`
}

type Host struct {
	VmPrefix          model.VmPrefix       `json:"vm-prefix"`
	VolumePrefix      model.VolumePrefix   `json:"disk-prefix"`
	LvmVolumeGroup    model.LvmVolumeGroup `json:"lvm-vg"`
	Name              model.HostName       `json:"name"`
	HostNetwork       string               `json:"host-network"`
	HostDevice        model.Device         `json:"host-device"`
	BackupNetwork     string               `json:"backup-network"`
	BackupDevice      model.Device         `json:"backup-device"`
	KubernetesNetwork string               `json:"kubernetes-network"`
	KubernetesDevice  model.Device         `json:"kubernetes-device"`
	KubernetesDns     string               `json:"kubernetes-dns"`
	Nodes             []Node               `json:"nodes"`
}

type Node struct {
	Name                model.NodeName `json:"name"`
	Master              bool           `json:"master"`
	Etcd                bool           `json:"etcd"`
	Worker              bool           `json:"worker"`
	Storage             bool           `json:"storage"`
	Nfsd                bool           `json:"nfsd"`
	Cores               int            `json:"cores"`
	Memory              int            `json:"memory"`
	Amount              int            `json:"number"`
	NfsSize             model.Size     `json:"nfssize"`
	StorageSize         model.Size     `json:"storagesize"`
	RootSize            model.Size     `json:"rootsize"`
	DockerSize          model.Size     `json:"dockersize"`
	KubeletSize         model.Size     `json:"kubeletsize"`
	Mac                 string         `json:"mac"`
	Address             string         `json:"address"`
	Gateway             string         `json:"gateway"`
	Dns                 string         `json:"dns"`
	ApiServerPort       int            `json:"apiserver-port"`
	IptablesFilterRules []string       `json:"iptables-filter-rules"`
	IptablesNatRules    []string       `json:"iptables-nat-rules"`
}
