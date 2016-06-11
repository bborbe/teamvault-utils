package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"text/template"

	"github.com/bborbe/kubernetes_tools/model"
	"github.com/bborbe/log"
	"os/user"
)

var logger = log.DefaultLogger

type generator struct {
}

type ConfigWriter interface {
	Write(config model.Cluster) error
}

func New() *generator {
	return new(generator)
}

func (c *generator) Write(cluster *model.Cluster) error {
	logger.Debugf("write config: %v", cluster)

	if err := createStructur(cluster); err != nil {
		return err
	}

	if err := writeUserDatas(cluster); err != nil {
		return err
	}

	if err := createScripts(cluster); err != nil {
		return err
	}

	return nil
}

func createStructur(cluster *model.Cluster) error {
	logger.Debugf("create user data")
	for _, node := range cluster.Nodes {

		if err := mkdir(fmt.Sprintf("%s/ssl", node.Name)); err != nil {
			return err
		}
		if err := touch(fmt.Sprintf("%s/ssl/.keep", node.Name)); err != nil {
			return err
		}
		if err := mkdir(fmt.Sprintf("%s/config/openstack/latest", node.Name)); err != nil {
			return err
		}
	}
	return nil
}


func createScripts(cluster *model.Cluster) error {
	logger.Debugf("create scripts")

	if err := mkdir("scripts"); err != nil {
		return err
	}

	if err := writeAdminCopyKeys(cluster); err != nil {
		return err
	}

	if err := writeAdminKubectlConfigure(cluster); err != nil {
		return err
	}

	if err := writeClusterCreate(cluster); err != nil {
		return err
	}

	if err := writeClusterDestroy(cluster); err != nil {
		return err
	}

	if err := writeStorageDataCreate(); err != nil {
		return err
	}

	if err := writeStorageDestroy(); err != nil {
		return err
	}

	if err := writeSSLCopyKeys(cluster); err != nil {
		return err
	}

	if err := writeSSLGenerateKeys(cluster); err != nil {
		return err
	}

	if err := writeMasterOpenssl(); err != nil {
		return err
	}

	if err := writeNodeOpenssl(); err != nil {
		return err
	}

	if err := writeVirshCreate(); err != nil {
		return err
	}

	if err := writeVirsh(cluster, "start"); err != nil {
		return err
	}

	if err := writeVirsh(cluster, "reboot"); err != nil {
		return err
	}

	if err := writeVirsh(cluster, "destroy"); err != nil {
		return err
	}

	if err := writeVirsh(cluster, "shutdown"); err != nil {
		return err
	}

	if err := writeVirsh(cluster, "undefine"); err != nil {
		return err
	}

	return nil
}

func writeUserDatas(cluster *model.Cluster) error {
	logger.Debugf("create user data")
	for _, node := range cluster.Nodes {
		if err := writeUserData(cluster, node); err != nil {
			return err
		}
	}
	return nil
}

func writeUserData(cluster *model.Cluster, node model.Node) error {
	logger.Debugf("write node %s", node.Name)

	var data struct {
		Name           string
		Region         string
		Mac            string
		Ip             string
		InitialCluster string
		EtcdEndpoints  string
		Etcd           bool
		Schedulable    bool
		Roles          string
		Nfsd           bool
		Storage        bool
		Master         bool
		ApiServers     string
		Gateway        string
		Dns            string
	}
	data.Name = node.Name
	data.Region = cluster.Region
	data.Mac = node.Mac
	data.Ip = node.Ip
	data.InitialCluster = cluster.InitialCluster()
	data.EtcdEndpoints = cluster.EtcdEndpoints()
	data.ApiServers = cluster.ApiServers()
	data.Etcd = node.Etcd
	data.Schedulable = node.Worker
	data.Roles = node.Roles()
	data.Nfsd = node.Storage
	data.Storage = node.Worker
	data.Master = node.Master
	data.Gateway = fmt.Sprintf("%s.1", cluster.Network)
	data.Dns = fmt.Sprintf("%s.1", cluster.Network)

	content, err := generateTemplate(`#cloud-config
ssh_authorized_keys:
 - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCOw/yh7+j3ygZp2aZRdZDWUh0Dkj5N9/USdiLSoS+0CHJta+mtSxxmI/yv1nOk7xnuA6qtjpxdMlWn5obtC9xyS6T++tlTK9gaPwU7a/PObtoZdfQ7znAJDpX0IPI06/OH1tFE9kEutHQPzhCwRaIQ402BHIrUMWzzP7Ige8Oa0HwXH4sHUG5h/V/svzi9T0CKJjF8dTx4iUfKX959hT8wQnKYPULewkNBFv6pNfWIr8EzvIEQcPmmm3tP+dQPKg5QKVi6jPdRla+t5HXfhXu0W3WCDa2s0VGmJjBdMMowr5MLNYI79MKziSV1w1IWL17Z58Lop0zEHqP7Ba0Aooqd
hostname: {{.Name}}
coreos:
  fleet:
    metadata: "region={{.Region}}"
  update:
    reboot-strategy: etcd-lock
  etcd2:
    name: "{{.Name}}"
    initial-cluster: "{{.InitialCluster}}"
    initial-cluster-token: "cluster-{{.Region}}"
{{if .Etcd}}
    initial-cluster-state: "new"
    initial-advertise-peer-urls: "http://{{.Ip}}:2380"
    advertise-client-urls: "http://{{.Ip}}:2379"
    listen-client-urls: "http://0.0.0.0:2379,http://0.0.0.0:4001"
    listen-peer-urls: "http://0.0.0.0:2380"
{{else}}
    listen-client-urls: "http://0.0.0.0:2379,http://0.0.0.0:4001"
    proxy: "on"
{{end}}
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
        MACAddress={{.Mac}}
        [Network]
        Address={{.Ip}}/24
        Gateway={{.Gateway}}
        DNS={{.Dns}}
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
{{if .Nfsd}}
    - name: data.mount
      command: start
      content: |
        [Unit]
        Description=Mount /data
        [Mount]
        What=/dev/vdc
        Where=/data
        Type=ext4
{{end}}
{{if .Storage}}
    - name: storage.mount
      command: start
      content: |
        [Unit]
        Description=Mount Storage to /storage
        [Mount]
        What=/dev/vdc
        Where=/storage
        Type=xfs
{{end}}
    - name: rpc-statd.service
      command: start
      enable: true
    - name: etcd2.service
      command: start
{{if .Nfsd}}
    - name: rpc-mountd.service
      command: start
    - name: nfsd.service
      command: start
{{end}}
    - name: fleet.service
      command: start
    - name:  systemd-networkd.service
      command: restart
    - name: flanneld.service
      command: start
{{if .Master}}
      drop-ins:
        - name: 50-network-config.conf
          content: |
            [Service]
            ExecStartPre=/usr/bin/etcdctl set /coreos.com/network/config '{ "Network": "10.102.0.0/16", "Backend":{"Type":"vxlan"} }'
{{end}}
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
{{if .Master}}
          --volume=/etc/kubernetes:/etc/kubernetes \
          --volume=/srv/kubernetes:/srv/kubernetes \
{{else}}
          --volume=/etc/kubernetes:/etc/kubernetes:ro \
{{end}}
          --net=host \
          --privileged=true \
          --pid=host \
          gcr.io/google_containers/hyperkube-amd64:v1.2.4 \
          /hyperkube kubelet \
            --containerized \
{{if .Master}}
            --api_servers=http://127.0.0.1:8080 \
{{else}}
            --api_servers={{.ApiServers}} \
{{end}}
            --register-node=true \
{{if not .Schedulable}}
            --register-schedulable=false \
{{end}}
            --allow-privileged=true \
            --config=/etc/kubernetes/manifests \
            --hostname-override={{.Ip}} \
            --cluster-dns=10.103.0.10 \
            --cluster-domain=cluster.local \
{{if not .Master}}
            --kubeconfig=/etc/kubernetes/node-kubeconfig.yaml \
            --tls-cert-file=/etc/kubernetes/ssl/node.pem \
            --tls-private-key-file=/etc/kubernetes/ssl/node-key.pem \
{{end}}
            --node-labels={{.Roles}} \
            --v=2
        [Install]
        WantedBy=multi-user.target
{{if .Master}}
    - name: kube-system-namespace.service
      command: start
      content: |
        [Unit]
        Description=Create Kube-System Namespace
        Requires=kubelet.service
        After=kubelet.service
        [Service]
        Restart=on-failure
        RestartSec=60s
        ExecStart=/bin/bash -c '\
          echo "try create namepsace kube-system"; \
          while true; do \
            curl -sf "http://127.0.0.1:8080/version"; \
            if [ $? -eq 0 ]; then \
              echo "api up. create namepsace kube-system"; \
              curl -XPOST -H Content-Type: application/json -d\'{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system"}}\' "http://127.0.0.1:8080/api/v1/namespaces"; \
              echo "namepsace kube-system created"; \
              exit 0; \
            else \
              echo "api down. sleep."; \
              sleep 20; \
            fi; \
          done'
        [Install]
        WantedBy=multi-user.target
{{end}}
write_files:
  - path: /etc/environment
    permissions: 0644
    content: |
      COREOS_PUBLIC_IPV4={{.Ip}}
      COREOS_PRIVATE_IPV4={{.Ip}}
  - path: /run/flannel/options.env
    permissions: 0644
    content: |
      FLANNELD_IFACE={{.Ip}}
      FLANNELD_ETCD_ENDPOINTS={{.EtcdEndpoints}}
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
{{if .Nfsd}}
  - path: /etc/exports
    permissions: 0644
    content: |
      /data/ 172.16.30.0/24(rw,async,no_subtree_check,no_root_squash,fsid=0)
{{end}}
{{if .Master}}
  - path: /etc/kubernetes/manifests/kube-apiserver.yaml
    permissions: 0644
    content: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: kube-apiserver
        namespace: kube-system
      spec:
        hostNetwork: true
        containers:
        - name: kube-apiserver
          image: gcr.io/google_containers/hyperkube-amd64:v1.2.4
          command:
          - /hyperkube
          - apiserver
          - --bind-address=0.0.0.0
          - --etcd-servers=http://172.16.30.15:2379,http://172.16.30.16:2379,http://172.16.30.17:2379
          - --allow-privileged=true
          - --service-cluster-ip-range=10.103.0.0/16
          - --secure-port=443
          - --advertise-address={{.Ip}}
          - --admission-control=NamespaceLifecycle,NamespaceExists,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota
          - --tls-cert-file=/etc/kubernetes/ssl/apiserver.pem
          - --tls-private-key-file=/etc/kubernetes/ssl/apiserver-key.pem
          - --client-ca-file=/etc/kubernetes/ssl/ca.pem
          - --service-account-key-file=/etc/kubernetes/ssl/apiserver-key.pem
          ports:
          - containerPort: 443
            hostPort: 443
            name: https
          - containerPort: 8080
            hostPort: 8080
            name: local
          volumeMounts:
          - mountPath: /etc/kubernetes/ssl
            name: ssl-certs-kubernetes
            readOnly: true
          - mountPath: /etc/ssl/certs
            name: ssl-certs-host
            readOnly: true
        volumes:
        - hostPath:
            path: /etc/kubernetes/ssl
          name: ssl-certs-kubernetes
        - hostPath:
            path: /usr/share/ca-certificates
          name: ssl-certs-host
  - path: /etc/kubernetes/manifests/kube-podmaster.yaml
    permissions: 0644
    content: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: kube-podmaster
        namespace: kube-system
      spec:
        hostNetwork: true
        containers:
        - name: controller-manager-elector
          image: gcr.io/google_containers/podmaster:1.1
          command:
          - /podmaster
          - --etcd-servers=http://172.16.30.15:2379,http://172.16.30.16:2379,http://172.16.30.17:2379
          - --key=controller
          - --whoami={{.Ip}}
          - --source-file=/src/manifests/kube-controller-manager.yaml
          - --dest-file=/dst/manifests/kube-controller-manager.yaml
          terminationMessagePath: /dev/termination-log
          volumeMounts:
          - mountPath: /src/manifests
            name: manifest-src
            readOnly: true
          - mountPath: /dst/manifests
            name: manifest-dst
        - name: scheduler-elector
          image: gcr.io/google_containers/podmaster:1.1
          command:
          - /podmaster
          - --etcd-servers=http://172.16.30.15:2379,http://172.16.30.16:2379,http://172.16.30.17:2379
          - --key=scheduler
          - --whoami={{.Ip}}
          - --source-file=/src/manifests/kube-scheduler.yaml
          - --dest-file=/dst/manifests/kube-scheduler.yaml
          volumeMounts:
          - mountPath: /src/manifests
            name: manifest-src
            readOnly: true
          - mountPath: /dst/manifests
            name: manifest-dst
        volumes:
        - hostPath:
            path: /srv/kubernetes/manifests
          name: manifest-src
        - hostPath:
            path: /etc/kubernetes/manifests
          name: manifest-dst
{{else}}
  - path: /etc/kubernetes/node-kubeconfig.yaml
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
{{end}}
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
{{if .Master}}
          - --master=http://127.0.0.1:8080
{{else}}
          - --master={{.ApiServers}}
          - --kubeconfig=/etc/kubernetes/node-kubeconfig.yaml
{{end}}
          - --proxy-mode=iptables
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /etc/ssl/certs
              name: ssl-certs-host
              readOnly: true
{{if not .Master}}
            - mountPath: /etc/kubernetes/node-kubeconfig.yaml
              name: "kubeconfig"
              readOnly: true
            - mountPath: /etc/kubernetes/ssl
              name: "etc-kube-ssl"
              readOnly: true
{{end}}
        volumes:
          - name: ssl-certs-host
            hostPath:
              path: "/usr/share/ca-certificates"
{{if not .Master}}
          - name: "kubeconfig"
            hostPath:
              path: "/etc/kubernetes/node-kubeconfig.yaml"
          - name: "etc-kube-ssl"
            hostPath:
              path: "/etc/kubernetes/ssl"
{{end}}
{{if .Master}}
  - path: /srv/kubernetes/manifests/kube-scheduler.yaml
    permissions: 0644
    content: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: kube-scheduler
        namespace: kube-system
      spec:
        hostNetwork: true
        containers:
        - name: kube-scheduler
          image: gcr.io/google_containers/hyperkube-amd64:v1.2.4
          command:
          - /hyperkube
          - scheduler
          - --master=http://127.0.0.1:8080
          livenessProbe:
            httpGet:
              host: 127.0.0.1
              path: /healthz
              port: 10251
            initialDelaySeconds: 15
            timeoutSeconds: 1
  - path: /srv/kubernetes/manifests/kube-controller-manager.yaml
    permissions: 0644
    content: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: kube-controller-manager
        namespace: kube-system
      spec:
        hostNetwork: true
        containers:
        - name: kube-controller-manager
          image: gcr.io/google_containers/hyperkube-amd64:v1.2.4
          command:
          - /hyperkube
          - controller-manager
          - --master=http://127.0.0.1:8080
          - --service-account-private-key-file=/etc/kubernetes/ssl/apiserver-key.pem
          - --root-ca-file=/etc/kubernetes/ssl/ca.pem
          livenessProbe:
            httpGet:
              host: 127.0.0.1
              path: /healthz
              port: 10252
            initialDelaySeconds: 15
            timeoutSeconds: 1
          volumeMounts:
            - mountPath: /etc/kubernetes/ssl
            name: ssl-certs-kubernetes
            readOnly: true
          - mountPath: /etc/ssl/certs
            name: ssl-certs-host
              readOnly: true
        volumes:
        - hostPath:
            path: /etc/kubernetes/ssl
          name: ssl-certs-kubernetes
        - hostPath:
            path: /usr/share/ca-certificates
          name: ssl-certs-host
{{end}}
`, data)
	if err != nil {
		return err
	}
	regex, err := regexp.Compile("\n+")
	if err != nil {
		return err
	}
	userData := []byte(regex.ReplaceAllString(string(content), "\n"))
	if err := writeFile(fmt.Sprintf("%s/config/openstack/latest/user_data", node.Name), userData, false); err != nil {
		return err
	}
	return nil
}

func writeAdminCopyKeys(cluster *model.Cluster) error {

	var data struct {
		Host   string
		Region string
		User   string
	}
	data.Host = cluster.Host
	data.Region = cluster.Region
	user, err := user.Current()
	if err != nil {
		return err
	}
	data.User = user.Username

	return writeTemplate("scripts/admin-copy-keys.sh", `#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

mkdir -p ~/.kube/{{.Region}}

scp {{.User}}@{{.Host}}:/var/lib/libvirt/images/kubernetes/scripts/kubernetes-ca.pem ~/.kube/{{.Region}}/
scp {{.User}}@{{.Host}}:/var/lib/libvirt/images/kubernetes/scripts/kubernetes-admin.pem ~/.kube/{{.Region}}/
scp {{.User}}@{{.Host}}:/var/lib/libvirt/images/kubernetes/scripts/kubernetes-admin-key.pem ~/.kube/{{.Region}}/
`, data, true)
}

func writeAdminKubectlConfigure(cluster *model.Cluster) error {

	var data struct {
		Region   string
		PublicIp string
	}
	data.Region = cluster.Region
	data.PublicIp = cluster.PublicIp

	return writeTemplate("scripts/admin-kubectl-configure.sh", `#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

mkdir -p $HOME/.kube/{{.Region}}
kubectl config set-cluster {{.Region}}-cluster --server=https://{{.PublicIp}}:443 --certificate-authority=$HOME/.kube/{{.Region}}/kubernetes-ca.pem
kubectl config set-credentials {{.Region}}-admin --certificate-authority=$HOME/.kube/{{.Region}}/kubernetes-ca.pem --client-key=$HOME/.kube/{{.Region}}/kubernetes-admin-key.pem --client-certificate=$HOME/.kube/{{.Region}}/kubernetes-admin.pem
kubectl config set-context {{.Region}}-system --cluster={{.Region}}-cluster --user={{.Region}}-admin
kubectl config use-context {{.Region}}-system

echo "test with 'kubectl get nodes'"
`, data, true)
}

func writeClusterCreate(cluster *model.Cluster) error {

	var data struct {
		VolumeNames []string
		VolumeGroup string
	}
	data.VolumeGroup = cluster.LvmVolumeGroup
	data.VolumeNames = cluster.VolumeNames()

	return writeTemplate("scripts/cluster-create.sh", `#!/bin/bash
{{$out := .}}
set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

echo "downloading image ..."
wget http://stable.release.core-os.net/amd64-usr/current/coreos_production_qemu_image.img.bz2 -O - | bzcat > /var/lib/libvirt/images/coreos_production_qemu_image.img
#wget http://beta.release.core-os.net/amd64-usr/current/coreos_production_qemu_image.img.bz2 -O - | bzcat > /var/lib/libvirt/images/coreos_production_qemu_image.img
#wget http://alpha.release.core-os.net/amd64-usr/current/coreos_production_qemu_image.img.bz2 -O - | bzcat > /var/lib/libvirt/images/coreos_production_qemu_image.img

echo "converting image ..."
qemu-img convert /var/lib/libvirt/images/coreos_production_qemu_image.img -O raw /var/lib/libvirt/images/coreos_production_qemu_image.raw

echo "create lvm volumes ..."
{{range $volumeName := .VolumeNames}}
lvcreate -L 10G -n {{$volumeName}} {{$out.VolumeGroup}}
lvcreate -L 10G -n {{$volumeName}}-docker {{$out.VolumeGroup}}
{{end}}

echo "writing images ..."
{{range $volumeName := .VolumeNames}}
dd bs=1M iflag=direct oflag=direct if=/var/lib/libvirt/images/coreos_production_qemu_image.raw of=/dev/{{$out.VolumeGroup}}/{{$volumeName}}
{{end}}

echo "cleanup"
rm /var/lib/libvirt/images/coreos_production_qemu_image.img /var/lib/libvirt/images/coreos_production_qemu_image.raw

${SCRIPT_ROOT}/virsh-create.sh

echo "done"
`, data, true)
}

func writeClusterDestroy(cluster *model.Cluster) error {

	var data struct {
		VolumeNames []string
		VolumeGroup string
	}
	data.VolumeGroup = cluster.LvmVolumeGroup
	data.VolumeNames = cluster.VolumeNames()

	return writeTemplate("scripts/cluster-destroy.sh", `#!/bin/bash
{{$out := .}}
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

${SCRIPT_ROOT}/virsh-destroy.sh
${SCRIPT_ROOT}/virsh-undefine.sh

echo "remove lvm volumes ..."
{{range $volumeName := .VolumeNames}}
lvremove /dev/{{$out.VolumeGroup}}/{{$volumeName}}
lvremove /dev/{{$out.VolumeGroup}}/{{$volumeName}}-docker
{{end}}

echo "done"
`, data, true)

}

func writeStorageDataCreate() error {

	var data struct{}

	return writeTemplate("scripts/storage-data-create.sh", `#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

echo "create lvm data volumes ..."
lvcreate -L 5G -n kubernetes-storage-data system

echo "format data volum ..."
wipefs /dev/system/kubernetes-storage-data
mkfs.ext4 -F /dev/system/kubernetes-storage-data

function create_storage {
	name="$1"
	echo "create lvm data volumes for ${name}"
	lvcreate -L 5G -n kubernetes-${name}-storage system

	echo "format data volum ..."
	wipefs /dev/system/kubernetes-${name}-storage
	mkfs.xfs -i size=512 /dev/system/kubernetes-${name}-storage
}

for ((i=0; i < 3; i++)) do
	create_storage "worker${i}"
done
`, data, true)
}

func writeStorageDestroy() error {

	var data struct{}

	return writeTemplate("scripts/storage-data-destroy.sh", `#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

lvremove /dev/system/kubernetes-storage-data

function delete_storage {
	name="$1"
	echo "remove lvm data volumes for worker ${name}"
	lvremove /dev/system/kubernetes-${name}-storage
}

for ((i=0; i < 3; i++)) do
	delete_storage "worker${i}"
done
`, data, true)
}

func writeSSLCopyKeys(cluster *model.Cluster) error {

	var data struct {
		NodeNames []string
	}
	data.NodeNames = cluster.NodeNames()

	return writeTemplate("scripts/ssl-copy-keys.sh", `#!/bin/bash
{{$out := .}}
set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

{{range $nodeName := .NodeNames}}
mkdir -p ${SCRIPT_ROOT}/../{{$nodeName}}/ssl
cp ${SCRIPT_ROOT}/kubernetes-ca.pem ${SCRIPT_ROOT}/../{{$nodeName}}/ssl/ca.pem
cp ${SCRIPT_ROOT}/{{$nodeName}}.pem ${SCRIPT_ROOT}/../{{$nodeName}}/ssl/node.pem
cp ${SCRIPT_ROOT}/{{$nodeName}}-key.pem ${SCRIPT_ROOT}/../{{$nodeName}}/ssl/node-key.pem
#chmod 600 ${SCRIPT_ROOT}/../{{$nodeName}}/ssl/*.pem
chown root:root ${SCRIPT_ROOT}/../{{$nodeName}}/ssl/*.pem
{{end}}
`, data, true)
}

func writeSSLGenerateKeys(cluster *model.Cluster) error {

	var data struct {
		PublicIp           string
		MasterNodeNames    []string
		NotMasterNodeNames []string
	}
	data.PublicIp = cluster.PublicIp
	data.MasterNodeNames = cluster.MasterNodeNames()
	data.NotMasterNodeNames = cluster.NotMasterNodeNames()

	return writeTemplate("scripts/ssl-generate-keys.sh", `#!/bin/bash
{{$out := .}}
set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

# https://coreos.com/kubernetes/docs/latest/openssl.html

# CA Key
openssl genrsa -out ${SCRIPT_ROOT}/kubernetes-ca-key.pem 2048
openssl req -x509 -new -nodes -key ${SCRIPT_ROOT}/kubernetes-ca-key.pem -days 10000 -out ${SCRIPT_ROOT}/kubernetes-ca.pem -subj "/CN=kube-ca"

# Admin Key
openssl genrsa -out ${SCRIPT_ROOT}/kubernetes-admin-key.pem 2048
openssl req -new -key ${SCRIPT_ROOT}/kubernetes-admin-key.pem -out ${SCRIPT_ROOT}/kubernetes-admin.csr -subj "/CN=kube-admin"
openssl x509 -req -in ${SCRIPT_ROOT}/kubernetes-admin.csr -CA ${SCRIPT_ROOT}/kubernetes-ca.pem -CAkey ${SCRIPT_ROOT}/kubernetes-ca-key.pem -CAcreateserial -out ${SCRIPT_ROOT}/kubernetes-admin.pem -days 365

{{range $nodeName := .MasterNodeNames}}
# {{$nodeName}}
openssl genrsa -out ${SCRIPT_ROOT}/kubernetes-apiserver-key.pem 2048
KUBERNETES_SVC=10.103.0.1 FIREWALL_IP={{$out.PublicIp}} MASTER_IP=${NODE_IP} openssl req -new -key ${SCRIPT_ROOT}/kubernetes-apiserver-key.pem -out ${SCRIPT_ROOT}/kubernetes-apiserver.csr -subj "/CN=kube-apiserver" -config ${SCRIPT_ROOT}/master-openssl.cnf
KUBERNETES_SVC=10.103.0.1 FIREWALL_IP={{$out.PublicIp}} MASTER_IP=${NODE_IP} openssl x509 -req -in ${SCRIPT_ROOT}/kubernetes-apiserver.csr -CA ${SCRIPT_ROOT}/kubernetes-ca.pem -CAkey ${SCRIPT_ROOT}/kubernetes-ca-key.pem -CAcreateserial -out ${SCRIPT_ROOT}/kubernetes-apiserver.pem -days 365 -extensions v3_req -extfile ${SCRIPT_ROOT}/master-openssl.cnf
{{end}}
{{range $nodeName := .NotMasterNodeNames}}
# {{$nodeName}}
openssl genrsa -out ${SCRIPT_ROOT}/kubernetes-${WORKER_FQDN}-key.pem 2048
NODE_IP=${NODE_IP} openssl req -new -key ${SCRIPT_ROOT}/kubernetes-${WORKER_FQDN}-key.pem -out ${SCRIPT_ROOT}/kubernetes-${WORKER_FQDN}.csr -subj "/CN=${WORKER_FQDN}" -config ${SCRIPT_ROOT}/node-openssl.cnf
NODE_IP=${NODE_IP} openssl x509 -req -in ${SCRIPT_ROOT}/kubernetes-${WORKER_FQDN}.csr -CA ${SCRIPT_ROOT}/kubernetes-ca.pem -CAkey ${SCRIPT_ROOT}/kubernetes-ca-key.pem -CAcreateserial -out ${SCRIPT_ROOT}/kubernetes-${WORKER_FQDN}.pem -days 365 -extensions v3_req -extfile ${SCRIPT_ROOT}/node-openssl.cnf
{{end}}
`, data, true)
}

func writeVirshCreate() error {

	var data struct{}

	return writeTemplate("scripts/virsh-create.sh", `#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})

function generate_mac {
	printf "00:16:3e:2f:20:%02x" $1
}

for ((i=0; i < 3; i++)) do
	NODEMAC=$(generate_mac $((15 + $i)))
	echo "create virsh kubernetes-etcd${i} mac=${NODEMAC} ..."
	virt-install \
	--import \
	--debug \
	--serial pty \
	--accelerate \
	--ram 750 \
	--vcpus 2 \
	--cpu=host \
	--os-type linux \
	--os-variant virtio26 \
	--noautoconsole \
	--nographics \
	--name kubernetes-etcd${i} \
	--disk /dev/system/kubernetes-etcd${i},bus=virtio,cache=none,io=native \
	--disk /dev/system/kubernetes-etcd${i}-docker,bus=virtio,cache=none,io=native \
	--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-etcd${i}/config/,config-2,type=mount,mode=squash \
	--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-etcd${i}/ssl/,kubernetes-ssl,type=mount,mode=squash \
	--network bridge=privatebr0,mac=${NODEMAC},model=virtio
done

NODEMAC=$(generate_mac "10")
echo "create virsh kubernetes-master mac=${NODEMAC} ..."
virt-install \
--import \
--debug \
--serial pty \
--accelerate \
--ram 1000 \
--vcpus 2 \
--cpu=host \
--os-type linux \
--os-variant virtio26 \
--noautoconsole \
--nographics \
--name kubernetes-master \
--disk /dev/system/kubernetes-master,bus=virtio,cache=none,io=native \
--disk /dev/system/kubernetes-master-docker,bus=virtio,cache=none,io=native \
--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-master/config/,config-2,type=mount,mode=squash \
--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-master/ssl/,kubernetes-ssl,type=mount,mode=squash \
--network bridge=privatebr0,mac=${NODEMAC},model=virtio

NODEMAC=$(generate_mac "9")
echo "create virsh kubernetes-storage mac=${NODEMAC} ..."
virt-install \
--import \
--debug \
--serial pty \
--accelerate \
--ram 750 \
--vcpus 2 \
--cpu=host \
--os-type linux \
--os-variant virtio26 \
--noautoconsole \
--nographics \
--name kubernetes-storage \
--disk /dev/system/kubernetes-storage,bus=virtio,cache=none,io=native \
--disk /dev/system/kubernetes-storage-docker,bus=virtio,cache=none,io=native \
--disk /dev/system/kubernetes-storage-data,bus=virtio,cache=none,io=native \
--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-storage/config/,config-2,type=mount,mode=squash \
--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-storage/ssl/,kubernetes-ssl,type=mount,mode=squash \
--network bridge=privatebr0,mac=${NODEMAC},model=virtio

for ((i=0; i < 3; i++)) do
	NODEMAC=$(generate_mac $((20 + $i)))
	echo "create virsh kubernetes-worker${i} mac=${NODEMAC} ..."
	virt-install \
	--import \
	--debug \
	--serial pty \
	--accelerate \
	--ram 3000 \
	--vcpus 2 \
	--cpu=host \
	--os-type linux \
	--os-variant virtio26 \
	--noautoconsole \
	--nographics \
	--name kubernetes-worker${i} \
	--disk /dev/system/kubernetes-worker${i},bus=virtio,cache=none,io=native \
	--disk /dev/system/kubernetes-worker${i}-docker,bus=virtio,cache=none,io=native \
	--disk /dev/system/kubernetes-worker${i}-storage,bus=virtio,cache=none,io=native \
	--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-worker${i}/config/,config-2,type=mount,mode=squash \
	--filesystem /var/lib/libvirt/images/kubernetes/kubernetes-worker${i}/ssl/,kubernetes-ssl,type=mount,mode=squash \
	--network bridge=privatebr0,mac=${NODEMAC},model=virtio
done
`, data, true)
}

func writeMasterOpenssl() error {

	var data struct{}

	return writeTemplate("scripts/master-openssl.cnf", `[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster.local
IP.1 = $ENV::KUBERNETES_SVC
IP.2 = $ENV::MASTER_IP
IP.3 = $ENV::FIREWALL_IP
`, data, false)
}

func writeNodeOpenssl() error {

	var script struct{}

	return writeTemplate("scripts/node-openssl.cnf", `[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names
[alt_names]
IP.1 = $ENV::NODE_IP
`, script, false)
}

func writeVirsh(cluster *model.Cluster, action string) error {
	var data struct {
		Action  string
		VmNames []string
	}
	data.Action = action
	data.VmNames = cluster.NodeNames()
	if err := writeTemplate(fmt.Sprintf("scripts/virsh-%s.sh", action), `#!/bin/bash
{{$out := .}}
set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

{{range $vmname := .VmNames}}
virsh {{$out.Action}} {{$vmname}}
{{end}}
`, data, true); err != nil {
		return err
	}
	return nil
}

func writeFile(path string, content []byte, executable bool) error {
	var perm os.FileMode
	if executable {
		perm = 0755
	} else {
		perm = 0644
	}
	return ioutil.WriteFile(path, content, perm)
}


func writeTemplate(path string, templateContent string, data interface{}, executable bool) error {
	content, err := generateTemplate(templateContent, data)
	if err != nil {
		return err
	}
	return writeFile(path, content, executable)
}

func generateTemplate(templateContent string, data interface{}) ([]byte, error) {
	tmpl, err := template.New("test").Parse(templateContent)
	if err != nil {
		return nil, err
	}
	content := bytes.NewBufferString("")
	if err := tmpl.Execute(content, data); err != nil {
		return nil, err
	}
	return content.Bytes(), nil
}

func mkdir(path string) error {
	var perm os.FileMode = 0755
	return os.MkdirAll(path, perm)
}

func touch(path string) error {
	return writeFile(path, make([]byte, 0), false)
}
