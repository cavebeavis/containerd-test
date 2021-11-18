package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gocni "github.com/containerd/go-cni"
)

// If namespace (for instance, "t1") has not been created and therefore exists under folder
// /var/run/netns/ then this will error out. Needs this created with the CNI api or with:
// sudo ip netns add t1

func main() {
	id := "example"
	netns := "/var/run/netns/t1"

	// CNI allows multiple CNI configurations and the network interface
	// will be named by eth0, eth1, ..., ethN.
	ifPrefixName := "veth"
	defaultIfName := "veth0"

	// Initializes library
	l, err := gocni.New(
		// one for loopback network interface
		gocni.WithMinNetworkCount(2),
		gocni.WithPluginConfDir("/etc/cni/net.d"),
		gocni.WithPluginDir([]string{"/opt/cni/bin"}),
		//gocni.WithPluginDir([]string{"/home/bk/go/bin"}),
		// Sets the prefix for network interfaces, eth by default
		gocni.WithInterfacePrefix(ifPrefixName))
	if err != nil {
		log.Fatalf("failed to initialize cni library: %v", err)
	}

	// Load the cni configuration
	if err := l.Load(gocni.WithLoNetwork, gocni.WithDefaultConf); err != nil {
		log.Fatalf("failed to load cni configuration: %v", err)
	}

	//log.Println("nets", l.GetConfig().Networks)
	for _, ntwrk := range l.GetConfig().Networks {
		log.Println(ntwrk.Config)
	}

	// Setup network for namespace.
	labels := map[string]string{
		//"K8S_POD_NAMESPACE":          "namespace1",
		//"K8S_POD_NAME":               "pod1",
		//"K8S_POD_INFRA_CONTAINER_ID": id,
		// Plugin tolerates all Args embedded by unknown labels, like
		// K8S_POD_NAMESPACE/NAME/INFRA_CONTAINER_ID...
		"IgnoreUnknown": "1",
	}

	portMap := []gocni.PortMapping{
		{
			HostPort:      26379,
			ContainerPort: 6379, // Redis for all!
			Protocol:      "tcp",
			HostIP:        "0.0.0.0",
		},
	}

	ctx := context.Background()

	// Teardown network
	defer func() {
		if err := l.Remove(ctx, id, netns, gocni.WithLabels(labels), gocni.WithCapabilityPortMap(portMap)); err != nil {
			log.Fatalf("failed to teardown network: %v", err)
		}
		fmt.Println("bye bye")
	}()

	// Setup network
	result, err := l.Setup(ctx, id, netns, gocni.WithLabels(labels), gocni.WithCapabilityPortMap(portMap))
	if err != nil {
		log.Fatalf("failed to setup network for namespace: %v", err)
	}

	log.Println(result.Interfaces)

	// Get IP of the default interface
	IP := result.Interfaces[defaultIfName].IPConfigs[0].IP.String()
	fmt.Printf("IP of the default interface %s:%s\n\n", defaultIfName, IP)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-shutdown:
		 fmt.Printf("\nkilled by: %s\n\n", sig.String())
		 return

	case <-time.After(time.Second * 120):
		break
	}
	
	fmt.Println("what? you want moreeee??? there is no more!!")
}