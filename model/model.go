package model

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

type UpdateRebootStrategy string
type KubernetesVersion string
type Region string
type VmPrefix string
type VolumePrefix string
type LvmVolumeGroup string
type Size string
type Device string
type HostName string

type Ip struct {
	ip net.IP
}

func (i *Ip) Set(pos int, value byte) {
	i.ip[pos] = value
}

func (i Ip) String() string {
	return i.ip.String()
}

func IpByString(ipString string) (*Ip, error) {
	ip := net.ParseIP(ipString)
	if len(ip) == 0 {
		return nil, fmt.Errorf("parse ip %s failed", ipString)
	}
	return &Ip{ip: ip}, nil
}

func (i Ip) Mac() (*Mac, error) {
	mac, err := MacByString("00:16:3e:75:cf:62")
	if err != nil {
		return nil, err
	}
	mac.mac[2] = i.ip[0]
	mac.mac[3] = i.ip[1]
	mac.mac[4] = i.ip[2]
	mac.mac[5] = i.ip[3]
	return mac, nil
}

type Gateway Ip

type Mac struct {
	mac net.HardwareAddr
}

func (m Mac) String() string {
	return m.mac.String()
}

func MacByString(macString string) (*Mac, error) {
	mac, err := net.ParseMAC(macString)
	if err != nil {
		return nil, err
	}
	return &Mac{mac}, nil
}

type Address struct {
	Ip   Ip
	Mask Mask
}

func (a Address) String() string {
	return fmt.Sprintf("%s/%d", a.Ip.String(), a.Mask)
}

type Dns Ip
type Mask int

type Cluster struct {
	Version              KubernetesVersion
	Region               Region
	UpdateRebootStrategy UpdateRebootStrategy
	Hosts                []Host
}

func (c *Cluster) Validate() error {
	if len(c.UpdateRebootStrategy) == 0 {
		return fmt.Errorf("Cluster.UpdateRebootStrategy missing")
	}
	if len(c.Region) == 0 {
		return fmt.Errorf("Cluster.Region missing")
	}
	if len(c.Version) == 0 {
		return fmt.Errorf("Cluster.Version missing")
	}
	if len(c.Hosts) == 0 {
		return fmt.Errorf("Cluster.Hosts missing")
	}
	for _, host := range c.Hosts {
		if err := host.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Host struct {
	Name           HostName
	LvmVolumeGroup LvmVolumeGroup
	VolumePrefix   VolumePrefix
	VmPrefix       VmPrefix
	Nodes          []Node
}

func (h *Host) Validate() error {
	if len(h.Name) == 0 {
		return fmt.Errorf("Host.Name missing")
	}
	if len(h.LvmVolumeGroup) == 0 {
		return fmt.Errorf("Host.LvmVolumeGroup missing")
	}
	for _, node := range h.Nodes {
		if err := node.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Node struct {
	HostNetwork      *Network
	KuberntesNetwork *Network
	BackupNetwork    *Network
	Name             string
	VolumeName       string
	VmName           string
	Etcd             bool
	Worker           bool
	Master           bool
	Nfsd             bool
	Storage          bool
	Cores            int
	Memory           int
	NfsSize          Size
	StorageSize      Size
	RootSize         Size
	DockerSize       Size
	KubeletSize      Size
}

func (n *Node) Validate() error {
	if len(n.Networks()) == 0 {
		return fmt.Errorf("Node.Networks missing")
	}
	if len(n.Name) == 0 {
		return fmt.Errorf("Node.Name missing")
	}
	if len(n.VolumeName) == 0 {
		return fmt.Errorf("Node.VolumeName missing")
	}
	if len(n.VmName) == 0 {
		return fmt.Errorf("Node.VmName missing")
	}
	if n.Cores <= 0 {
		return fmt.Errorf("Node.Cores missing")
	}
	if n.Memory <= 0 {
		return fmt.Errorf("Node.Memory missing")
	}
	return nil
}

func (n *Node) Networks() []Network {
	result := []Network{}
	if n.HostNetwork != nil {
		result = append(result, *n.HostNetwork)
	}
	if n.KuberntesNetwork != nil {
		result = append(result, *n.KuberntesNetwork)
	}
	if n.BackupNetwork != nil {
		result = append(result, *n.BackupNetwork)
	}
	return result
}

type Network struct {
	Device  Device
	Mac     Mac
	Address Address
	Gateway Gateway
	Dns     Dns
}

func (c *Host) VolumeNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.VolumeName)
	}
	return result
}

func (c *Host) NodeNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.Name)
	}
	return result
}

func (c *Host) VmNames() []string {
	var result []string
	for _, node := range c.Nodes {
		result = append(result, node.VmName)
	}
	return result
}

func (c *Host) MasterNodes() []Node {
	var result []Node
	for _, node := range c.Nodes {
		if node.Master {
			result = append(result, node)
		}
	}
	return result
}

func (c *Host) NotMasterNodes() []Node {
	var result []Node
	for _, node := range c.Nodes {
		if !node.Master {
			result = append(result, node)
		}
	}
	return result
}

func (c *Host) StorageNodes() []Node {
	var result []Node
	for _, node := range c.Nodes {
		if node.Storage {
			result = append(result, node)
		}
	}
	return result
}

func (c *Host) NfsdNodes() []Node {
	var result []Node
	for _, node := range c.Nodes {
		if node.Nfsd {
			result = append(result, node)
		}
	}
	return result
}

func (c *Host) EtcdEndpoints() string {
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
			content.WriteString(node.KuberntesNetwork.Address.Ip.String())
			content.WriteString(":2379")
		}
	}
	return content.String()
}

func (c *Host) InitialCluster() string {
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
			content.WriteString(node.KuberntesNetwork.Address.Ip.String())
			content.WriteString(":2380")
		}
	}
	return content.String()
}

func (c *Host) ApiServers() string {
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
			content.WriteString(node.KuberntesNetwork.Address.Ip.String())
		}
	}
	return content.String()
}

func (n *Node) Labels() string {
	var labels []string
	if n.Etcd {
		labels = append(labels, "etcd=true")
	}
	if n.Storage {
		labels = append(labels, "storage=true")
	}
	if n.Nfsd {
		labels = append(labels, "nfsd=true")
	}
	if n.Worker {
		labels = append(labels, "worker=true")
	}
	if n.Master {
		labels = append(labels, "master=true")
	}
	return strings.Join(labels, ",")
}
