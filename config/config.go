package config

type Cluster struct {
	Region         string `json:"region"`
	PublicIp       string `json:"public-ip"`
	Network        string `json:"network"`
	MacPrefix      string `json:"macprefix"`
	LvmVolumeGroup string `json:"lvm-vg"`
	VmPrefix       string `json:"vm-prefix"`
	VolumePrefix   string `json:"disk-prefix"`
	Host           string `json:"host"`
	Bridge         string `json:"bridge"`
	Nodes          []Node `json:"nodes"`
}

type Node struct {
	Name    string `json:"name"`
	Etcd    bool   `json:"etcd"`
	Worker  bool   `json:"worker"`
	Storage bool   `json:"nfsd"`
	Master  bool   `json:"master"`
	Cores   int    `json:"cores"`
	Amount  int    `json:"number"`
}
