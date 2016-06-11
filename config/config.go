package config

type Cluster struct {
	Region            string `json:"region"`
	ApiServerPublicIp string `json:"api-server-ip"`
	Network           string `json:"network"`
	MacPrefix         string `json:"macprefix"`
	LvmVolumeGroup    string `json:"lvm-vg"`
	VmPrefix          string `json:"vm-prefix"`
	VolumePrefix      string `json:"disk-prefix"`
	Host              string `json:"host"`
	Bridge            string `json:"bridge"`
	Nodes             []Node `json:"nodes"`
}

type Node struct {
	Name        string `json:"name"`
	Master      bool   `json:"master"`
	Etcd        bool   `json:"etcd"`
	Worker      bool   `json:"worker"`
	Storage     bool   `json:"storage"`
	Nfsd        bool   `json:"nfsd"`
	Cores       int    `json:"cores"`
	Memory      int    `json:"memory"`
	Amount      int    `json:"number"`
	NfsSize     string `json:"nfssize"`
	StorageSize string `json:"storagesize"`
}
