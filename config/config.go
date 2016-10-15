package config

import "github.com/bborbe/kubernetes_tools/model"

type Cluster struct {
	UpdateRebootStrategy model.UpdateRebootStrategy `json:"update-reboot-strategy"`
	KubernetesVersion    model.KubernetesVersion    `json:"kuberntes-version"`
	Region               model.Region               `json:"region"`
	Hosts                []Host                     `json:"hosts"`
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
	Nodes             []Node               `json:"nodes"`
}

type Node struct {
	Name        model.NodeName `json:"name"`
	Master      bool           `json:"master"`
	Etcd        bool           `json:"etcd"`
	Worker      bool           `json:"worker"`
	Storage     bool           `json:"storage"`
	Nfsd        bool           `json:"nfsd"`
	Cores       int            `json:"cores"`
	Memory      int            `json:"memory"`
	Amount      int            `json:"number"`
	NfsSize     model.Size     `json:"nfssize"`
	StorageSize model.Size     `json:"storagesize"`
	RootSize    model.Size     `json:"rootsize"`
	DockerSize  model.Size     `json:"dockersize"`
	KubeletSize model.Size     `json:"kubeletsize"`
}
