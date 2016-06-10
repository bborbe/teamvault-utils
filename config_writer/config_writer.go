package config_writer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"text/template"

	"github.com/bborbe/kubernetes_tools/config"
	"github.com/bborbe/log"
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

	counter := 0
	for _, node := range cluster.Nodes {
		for i := 0; i < node.Number; i++ {
			counter++
			logger.Debugf("generate node %d started", counter)
			if err := writeNode(cluster, node, i, counter); err != nil {
				return err
			}
			logger.Debugf("generate node %d finished", counter)
		}
	}

	return nil
}

func writeNode(cluster config.Cluster, node config.Node, number int, counter int) error {
	name := generateNodeName(node, number)
	logger.Debugf("write node %s", name)

	var configuration NodeConfiguration
	configuration.Name = name
	configuration.Mac = generateMac(cluster.MacPrefix, counter)
	configuration.Ip = generateIp(cluster.Network, counter)
	configuration.InitialCluster = generateInitialCluster(cluster)
	configuration.EtcdEndpoints = generateEtcdEndpoints(cluster)
	configuration.ApiServers = generateApiServers(cluster)
	configuration.Etcd = node.Etcd
	configuration.Schedulable = node.Worker
	configuration.Roles = generateRoles(node)
	configuration.Nfsd = node.Storage
	configuration.Storage = node.Worker
	configuration.Master = node.Master

	if err := createClusterConfig(configuration); err != nil {
		return err
	}
	return nil
}

func generateRoles(node config.Node) string {
	var roles []string
	if node.Etcd {
		roles = append(roles, "etcd")
	}
	if node.Worker {
		roles = append(roles, "worker")
	}
	if node.Master {
		roles = append(roles, "master")
	}
	if node.Storage {
		roles = append(roles, "storage")
	}
	content := bytes.NewBufferString("")
	for i, role := range roles {
		if i != 0 {
			content.WriteString(",")
		}
		content.WriteString("role=")
		content.WriteString(role)
	}
	return content.String()
}

func generateNodeName(node config.Node, number int) string {
	if node.Number == 1 {
		return node.Name
	} else {
		return fmt.Sprintf("%s%d", node.Name, number)
	}
}

func generateApiServers(cluster config.Cluster) string {
	first := true
	content := bytes.NewBufferString("")
	counter := 0
	for _, node := range cluster.Nodes {
		for i := 0; i < node.Number; i++ {
			counter++
			if node.Master {
				if first {
					first = false
				} else {
					content.WriteString(",")
				}
				ip := generateIp(cluster.Network, counter)
				content.WriteString("https://")
				content.WriteString(ip)
			}
		}
	}
	return content.String()
}

func generateInitialCluster(cluster config.Cluster) string {
	first := true
	content := bytes.NewBufferString("")
	counter := 0
	for _, node := range cluster.Nodes {
		for i := 0; i < node.Number; i++ {
			counter++
			if node.Etcd {
				if first {
					first = false
				} else {
					content.WriteString(",")
				}
				name := generateNodeName(node, i)
				ip := generateIp(cluster.Network, counter)
				content.WriteString(name)
				content.WriteString("=http://")
				content.WriteString(ip)
				content.WriteString(":2380")
			}
		}
	}
	return content.String()
}

func generateEtcdEndpoints(cluster config.Cluster) string {
	first := true
	content := bytes.NewBufferString("")
	counter := 0
	for _, node := range cluster.Nodes {
		for i := 0; i < node.Number; i++ {
			counter++
			if node.Etcd {
				if first {
					first = false
				} else {
					content.WriteString(",")
				}
				ip := generateIp(cluster.Network, counter)
				content.WriteString("http://")
				content.WriteString(ip)
				content.WriteString(":2379")
			}
		}
	}
	return content.String()
}

func generateMac(prefix string, counter int) string {
	return fmt.Sprintf("%s%02x", prefix, counter+10)
}

func generateIp(prefix string, counter int) string {
	return fmt.Sprintf("%s.%d", prefix, counter+10)
}

func createClusterConfig(node NodeConfiguration) error {
	if err := mkdir(fmt.Sprintf("%s/ssl", node.Name)); err != nil {
		return err
	}
	if err := touch(fmt.Sprintf("%s/ssl/.keep", node.Name)); err != nil {
		return err
	}
	if err := mkdir(fmt.Sprintf("%s/config/openstack/latest", node.Name)); err != nil {
		return err
	}
	userData, err := generateUserDataContent(node)
	if err != nil {
		return err
	}
	if err := writeFile(fmt.Sprintf("%s/config/openstack/latest/user_data", node.Name), userData); err != nil {
		return err
	}
	return nil
}

func writeFile(path string, content []byte) error {
	var perm os.FileMode = 0644
	return ioutil.WriteFile(path, content, perm)
}

type NodeConfiguration struct {
	Name           string
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
}

func generateUserDataContent(userData NodeConfiguration) ([]byte, error) {
	tmpl, err := template.New("test").Parse(`#cloud-config
ssh_authorized_keys:
 - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCOw/yh7+j3ygZp2aZRdZDWUh0Dkj5N9/USdiLSoS+0CHJta+mtSxxmI/yv1nOk7xnuA6qtjpxdMlWn5obtC9xyS6T++tlTK9gaPwU7a/PObtoZdfQ7znAJDpX0IPI06/OH1tFE9kEutHQPzhCwRaIQ402BHIrUMWzzP7Ige8Oa0HwXH4sHUG5h/V/svzi9T0CKJjF8dTx4iUfKX959hT8wQnKYPULewkNBFv6pNfWIr8EzvIEQcPmmm3tP+dQPKg5QKVi6jPdRla+t5HXfhXu0W3WCDa2s0VGmJjBdMMowr5MLNYI79MKziSV1w1IWL17Z58Lop0zEHqP7Ba0Aooqd
hostname: {{.Name}}
coreos:
  fleet:
    metadata: "region=rn"
  update:
    reboot-strategy: etcd-lock
  etcd2:
    name: "{{.Name}}"
    initial-cluster: "{{.InitialCluster}}"
    initial-cluster-token: "cluster-rn"
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
          - --advertise-address=172.16.30.10
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
          - --whoami=172.16.30.10
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
          - --whoami=172.16.30.10
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
`)
	if err != nil {
		return nil, err
	}
	content := bytes.NewBufferString("")
	tmpl.Execute(content, userData)

	regex, err := regexp.Compile("\n+")
	if err != nil {
		return nil, err
	}
	return []byte(regex.ReplaceAllString(content.String(), "\n")), nil
}

func mkdir(path string) error {
	var perm os.FileMode = 0755
	return os.MkdirAll(path, perm)
}

func touch(path string) error {
	return writeFile(path, make([]byte, 0))
}
