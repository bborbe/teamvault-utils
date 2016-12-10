package model

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
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

func (d Device) String() string {
	return string(d)
}

type HostName string

type NodeName string

func (n NodeName) String() string {
	return string(n)
}

type VolumeName string

func (v VolumeName) String() string {
	return string(v)
}

type VmName string

func (v VmName) String() string {
	return string(v)
}

type Ip struct {
	ip net.IP
}

func (i *Ip) Set(pos int, value byte) {
	ip := net.IPv4(i.ip[12], i.ip[13], i.ip[14], i.ip[15])
	ip[pos] = value
	i.ip = ip
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
	mac.mac[2] = i.ip[12]
	mac.mac[3] = i.ip[13]
	mac.mac[4] = i.ip[14]
	mac.mac[5] = i.ip[15]
	return mac, nil
}

type Gateway Ip

func (g Gateway) String() string {
	return g.ip.String()
}

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

func (a *Address) Validate() error {
	if len(a.Ip.ip) == 4 {
		return fmt.Errorf("Address.Ip invalid")
	}
	if a.Mask == 0 {
		return fmt.Errorf("Address.Mask invalid")
	}
	return nil
}

func ParseAddress(address string) (*Address, error) {
	parts := strings.Split(address, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("parse address %v failed", address)
	}
	mask, err := ParseMask(parts[1])
	if err != nil {
		return nil, err
	}
	ip, err := IpByString(parts[0])
	if err != nil {
		return nil, err
	}
	return &Address{
		Ip:   *ip,
		Mask: *mask,
	}, nil
}

func (a Address) String() string {
	return fmt.Sprintf("%s/%d", a.Ip.String(), a.Mask)
}

func (a Address) Network() string {
	ipnet := net.IPNet{
		IP:   a.Ip.ip,
		Mask: net.CIDRMask(a.Mask.Int(), 8*net.IPv4len),
	}
	return fmt.Sprintf("%s/%d", ipnet.IP.Mask(ipnet.Mask).String(), a.Mask.Int())
}

type Dns Ip

func (d Dns) String() string {
	return d.ip.String()
}

type Mask int

func (m Mask) String() string {
	return strconv.Itoa(m.Int())
}

func (m Mask) Int() int {
	return int(m)
}
func ParseMask(mask string) (*Mask, error) {
	i, err := strconv.Atoi(mask)
	if err != nil {
		return nil, err
	}
	m := Mask(i)
	return &m, nil
}

type Cluster struct {
	Version              KubernetesVersion
	Region               Region
	UpdateRebootStrategy UpdateRebootStrategy
	Hosts                []Host
}

func (c *Cluster) Validate(features Features) error {
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
		if err := host.Validate(features); err != nil {
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

func (h *Host) Validate(features Features) error {
	if len(h.Name) == 0 {
		return fmt.Errorf("Host.Name missing")
	}
	if features.Kvm && len(h.LvmVolumeGroup) == 0 {
		return fmt.Errorf("Host.LvmVolumeGroup missing")
	}
	for _, node := range h.Nodes {
		if err := node.Validate(features); err != nil {
			return err
		}
	}
	return nil
}

type Node struct {
	Name              NodeName
	HostNetwork       *Network
	KubernetesNetwork *Network
	BackupNetwork     *Network
	VolumeName        VolumeName
	VmName            VmName
	Etcd              bool
	Worker            bool
	Master            bool
	Nfsd              bool
	Storage           bool
	Cores             int
	Memory            int
	NfsSize           Size
	StorageSize       Size
	RootSize          Size
	DockerSize        Size
	KubeletSize       Size
}

func (n *Node) Validate(features Features) error {
	if features.Kvm && len(n.Networks()) == 0 {
		return fmt.Errorf("Node.Networks missing")
	}
	for _, network := range n.Networks() {
		if err := network.Validate(features); err != nil {
			return err
		}
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
	if features.Kvm && n.Cores <= 0 {
		return fmt.Errorf("Node.Cores missing")
	}
	if features.Kvm && n.Memory <= 0 {
		return fmt.Errorf("Node.Memory missing")
	}
	return nil
}

func (n *Node) Networks() []Network {
	result := []Network{}
	if n.HostNetwork != nil {
		result = append(result, *n.HostNetwork)
	}
	if n.KubernetesNetwork != nil {
		result = append(result, *n.KubernetesNetwork)
	}
	if n.BackupNetwork != nil {
		result = append(result, *n.BackupNetwork)
	}
	return result
}

type Network struct {
	Number  int
	Device  Device
	Mac     Mac
	Address Address
	Gateway Gateway
	Dns     Dns
}

func (n *Network) Validate(features Features) error {
	if err := n.Address.Validate(); err != nil {
		return err
	}
	if !features.Kvm {
		return nil
	}
	if len(n.Device) == 0 {
		return fmt.Errorf("Network.Device missing")
	}
	if len(n.Mac.String()) == 0 {
		return fmt.Errorf("Network.Mac missing")
	}
	if len(n.Gateway.String()) == 0 {
		return fmt.Errorf("Network.Gateway missing")
	}
	if len(n.Dns.String()) == 0 {
		return fmt.Errorf("Network.Dns missing")
	}
	return nil
}

func (c *Host) VolumeNames() []VolumeName {
	var result []VolumeName
	for _, node := range c.Nodes {
		result = append(result, node.VolumeName)
	}
	return result
}

func (c *Host) NodeNames() []NodeName {
	var result []NodeName
	for _, node := range c.Nodes {
		result = append(result, node.Name)
	}
	return result
}

func (c *Host) VmNames() []VmName {
	var result []VmName
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
			content.WriteString(node.KubernetesNetwork.Address.Ip.String())
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
			content.WriteString(node.Name.String())
			content.WriteString("=http://")
			content.WriteString(node.KubernetesNetwork.Address.Ip.String())
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
			content.WriteString(node.KubernetesNetwork.Address.Ip.String())
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

type Features struct {
	Kvm bool
}
