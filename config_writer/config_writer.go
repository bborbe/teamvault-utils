package config_writer

import (
	"github.com/bborbe/kubernetes_tools/config"
	"github.com/bborbe/log"
	"fmt"
	"os"
	"io/ioutil"
	"bytes"
)

var logger = log.DefaultLogger

type configWriter struct {

}

type ConfigWriter interface {
	WriteConfigs(config config.Cluster) error
}

func New() *configWriter {
	return new(configWriter)
}

func (c *configWriter) WriteConfigs(cluster config.Cluster) error {
	logger.Debugf("write config: %v", cluster)

	for _, node := range cluster.Nodes {
		if err := writeNode(cluster, node); err != nil {
			return err
		}
	}

	return nil
}

func writeNode(cluster config.Cluster, node config.Node) error {
	for i := 0; i < node.Number; i++ {
		var name string
		if (node.Number == 1) {
			name = node.Name
		} else {
			name = fmt.Sprintf("%s%d", node.Name, i)
		}
		if err := createClusterConfig(name); err != nil {
			return err
		}
	}
	return nil
}

func createClusterConfig(name string) error {
	if err := mkdir(fmt.Sprintf("%s/ssl", name)); err != nil {
		return err
	}
	if err := touch(fmt.Sprintf("%s/ssl/.keep", name)); err != nil {
		return err
	}
	if err := mkdir(fmt.Sprintf("%s/config/openstack/latest", name)); err != nil {
		return err
	}
	userData, err := generateUserDataContent()
	if err != nil {
		return err
	}
	if err := writeFile(fmt.Sprintf("%s/config/openstack/latest/user_data", name), userData); err != nil {
		return err
	}
	return nil
}

func writeFile(path string, content []byte) error {
	var perm os.FileMode = 0644
	return ioutil.WriteFile(path, content, perm)
}

func generateUserDataContent() ([]byte, error) {
	content := bytes.NewBufferString("")
	content.WriteString(`#cloud-config
ssh_authorized_keys:
 - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCOw/yh7+j3ygZp2aZRdZDWUh0Dkj5N9/USdiLSoS+0CHJta+mtSxxmI/yv1nOk7xnuA6qtjpxdMlWn5obtC9xyS6T++tlTK9gaPwU7a/PObtoZdfQ7znAJDpX0IPI06/OH1tFE9kEutHQPzhCwRaIQ402BHIrUMWzzP7Ige8Oa0HwXH4sHUG5h/V/svzi9T0CKJjF8dTx4iUfKX959hT8wQnKYPULewkNBFv6pNfWIr8EzvIEQcPmmm3tP+dQPKg5QKVi6jPdRla+t5HXfhXu0W3WCDa2s0VGmJjBdMMowr5MLNYI79MKziSV1w1IWL17Z58Lop0zEHqP7Ba0Aooqd
hostname: kubernetes-etcd0
coreos:
  fleet:
    metadata: "region=rn"
  update:
    reboot-strategy: etcd-lock
  etcd2:
    name: "kubernetes-etcd0"
    initial-cluster: "kubernetes-etcd0=http://172.16.30.15:2380,kubernetes-etcd1=http://172.16.30.16:2380,kubernetes-etcd2=http://172.16.30.17:2380"
    initial-cluster-token: "cluster-rn"
    initial-cluster-state: "new"
    initial-advertise-peer-urls: "http://172.16.30.15:2380"
    advertise-client-urls: "http://172.16.30.15:2379"
    listen-client-urls: "http://0.0.0.0:2379,http://0.0.0.0:4001"
    listen-peer-urls: "http://0.0.0.0:2380"
  units:
    - name: etc-kubernetes-ssl.mount
      command: start
      content: |
        [Unit]
        Wants=user-configvirtfs.service
        Before=user-configvirtfs.service
        # Only mount config drive block devices automatically in virtual machines
        # or any host that has it explicitly enabled and not explicitly disabled.
        ConditionVirtualization=|vm
        ConditionKernelCommandLine=|coreos.configdrive=1
        ConditionKernelCommandLine=!coreos.configdrive=0
        [Mount]
        What=kubernetes-ssl
        Where=/etc/kubernetes/ssl
        Options=ro,trans=virtio,version=9p2000.L
        Type=9p
    - name: 10-ens3.network
      content: |
        [Match]
        MACAddress=00:16:3e:2f:30:0f
        [Network]
        Address=172.16.30.15/24
        Gateway=172.16.30.1
        DNS=172.16.30.1
    - name: format-ephemeral.service
      command: start
      content: |
        [Unit]
        Description=Formats the ephemeral drive
        After=dev-vdb.device
        Requires=dev-vdb.device
        [Service]
        Type=oneshot
        RemainAfterExit=yes
        ExecStart=/usr/sbin/wipefs -f /dev/vdb
        ExecStart=/usr/sbin/mkfs.ext4 -i 4096 -F /dev/vdb
    - name: var-lib-docker.mount
      command: start
      content: |
        [Unit]
        Description=Mount /var/lib/docker
        Requires=format-ephemeral.service
        After=format-ephemeral.service
        [Mount]
        What=/dev/vdb
        Where=/var/lib/docker
        Type=ext4
    - name: rpc-statd.service
      command: start
      enable: true
    - name: etcd2.service
      command: start
    - name: fleet.service
      command: start
    - name:  systemd-networkd.service
      command: restart
    - name: flanneld.service
      command: start
    - name: docker.service
      command: start
      drop-ins:
        - name: 40-flannel.conf
          content: |
            [Unit]
            Requires=flanneld.service
            After=flanneld.service
        - name: 10-wait-docker.conf
          content: |
            [Unit]
            After=var-lib-docker.mount
            Requires=var-lib-docker.mount
    - name: docker-cleanup.service
      content: |
        [Unit]
        Description=Docker Cleanup
        Requires=docker.service
        After=docker.service
        [Service]
        Type=oneshot
        ExecStart=-/bin/bash -c '/usr/bin/docker rm -v $(/usr/bin/docker ps -a -q -f status=exited)'
        ExecStart=-/bin/bash -c '/usr/bin/docker rmi $(/usr/bin/docker images -f dangling=true -q)'
    - name: docker-cleanup.timer
      command: start
      content: |
        [Unit]
        Description=Docker Cleanup every 4 hours
        [Timer]
        Unit=docker-cleanup.service
        OnCalendar=*-*-* 0/4:00:00
        [Install]
        WantedBy=multi-user.target
    - name: kubelet.service
      command: start
      content: |
        [Unit]
        Description=Kubelet
        Requires=docker.service
        After=docker.service
        [Service]
        Restart=always
        RestartSec=20s
        EnvironmentFile=/etc/environment
        TimeoutStartSec=0
        ExecStart=/usr/bin/docker run \
          --volume=/:/rootfs:ro \
          --volume=/sys:/sys:ro \
          --volume=/var/lib/docker/:/var/lib/docker:rw \
          --volume=/var/lib/kubelet/:/var/lib/kubelet:rw \
          --volume=/var/run:/var/run:rw \
          --volume=/etc/kubernetes:/etc/kubernetes:ro \
          --net=host \
          --privileged=true \
          --pid=host \
          gcr.io/google_containers/hyperkube-amd64:v1.2.4 \
          /hyperkube kubelet \
            --containerized \
            --api_servers=https://172.16.30.10 \
            --register-node=true \
            --register-schedulable=false \
            --allow-privileged=true \
            --config=/etc/kubernetes/manifests \
            --hostname-override=172.16.30.15 \
            --cluster-dns=10.103.0.10 \
            --cluster-domain=cluster.local \
            --kubeconfig=/etc/kubernetes/storage-kubeconfig.yaml \
            --tls-cert-file=/etc/kubernetes/ssl/node.pem \
            --tls-private-key-file=/etc/kubernetes/ssl/node-key.pem \
            --node-labels=role=etcd \
            --v=2
        [Install]
        WantedBy=multi-user.target
write_files:
  - path: /etc/environment
    permissions: 0644
    content: |
      COREOS_PUBLIC_IPV4=172.16.30.15
      COREOS_PRIVATE_IPV4=172.16.30.15
  - path: /run/flannel/options.env
    permissions: 0644
    content: |
      FLANNELD_IFACE=172.16.30.15
      FLANNELD_ETCD_ENDPOINTS=http://172.16.30.15:2379,http://172.16.30.16:2379,http://172.16.30.17:2379
  - path: /root/.toolboxrc
    owner: core
    content: |
      TOOLBOX_DOCKER_IMAGE=bborbe/toolbox
      TOOLBOX_DOCKER_TAG=latest
      TOOLBOX_USER=root
  - path: /home/core/.toolboxrc
    owner: core
    content: |
      TOOLBOX_DOCKER_IMAGE=bborbe/toolbox
      TOOLBOX_DOCKER_TAG=latest
      TOOLBOX_USER=root
  - path: /etc/exports
    permissions: 0644
    content: |
      /data/ 172.16.30.0/24(rw,async,no_subtree_check,no_root_squash,fsid=0)
  - path: /etc/kubernetes/storage-kubeconfig.yaml
    permissions: 0644
    content: |
      apiVersion: v1
      kind: Config
      clusters:
      - name: local
        cluster:
          certificate-authority: /etc/kubernetes/ssl/ca.pem
      users:
      - name: kubelet
        user:
          client-certificate: /etc/kubernetes/ssl/node.pem
          client-key: /etc/kubernetes/ssl/node-key.pem
      contexts:
      - context:
          cluster: local
          user: kubelet
        name: kubelet-context
      current-context: kubelet-context
  - path: /etc/kubernetes/manifests/kube-proxy.yaml
    permissions: 0644
    content: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: kube-proxy
        namespace: kube-system
      spec:
        hostNetwork: true
        containers:
        - name: kube-proxy
          image: gcr.io/google_containers/hyperkube-amd64:v1.2.4
          command:
          - /hyperkube
          - proxy
          - --master=https://172.16.30.10
          - --kubeconfig=/etc/kubernetes/storage-kubeconfig.yaml
          - --proxy-mode=iptables
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /etc/ssl/certs
              name: "ssl-certs"
            - mountPath: /etc/kubernetes/storage-kubeconfig.yaml
              name: "kubeconfig"
              readOnly: true
            - mountPath: /etc/kubernetes/ssl
              name: "etc-kube-ssl"
              readOnly: true
        volumes:
          - name: "ssl-certs"
            hostPath:
              path: "/usr/share/ca-certificates"
          - name: "kubeconfig"
            hostPath:
              path: "/etc/kubernetes/storage-kubeconfig.yaml"
          - name: "etc-kube-ssl"
            hostPath:
              path: "/etc/kubernetes/ssl"
`)
	return content.Bytes(), nil
}

func mkdir(path string) error {
	var perm os.FileMode = 0755
	return os.MkdirAll(path, perm)
}

func touch(path string) error {
	return writeFile(path, make([]byte, 0))
}
