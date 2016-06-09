package config

type Cluster struct {
	Region      string `json:"region"`
	PublicIp    string `json:"public-ip"`
	Network     string `json:"network"`
	MacPrefix   string `json:"macprefix"`
	LvmVvg      string `json:"lvm-vg"`
	VmPrefix    string `json:"vm-prefix"`
	DiskPprefix string `json:"disk-prefix"`
	Host        string `json:"host"`
	Bridge      string `json:"bridge"`
	Nodes       []Node `json:"nodes"`
}

type Node struct {
	Name    string `json:"name"`
	Etcd    bool `json:"etcd"`
	Worker  bool `json:"worker"`
	Storage bool `json:"nfsd"`
	Master  bool `json:"master"`
	Cores   int `json:"cores"`
	Number  int `json:"number"`
}


