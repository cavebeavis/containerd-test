package main

import (
	"context"
	"log"
	"strings"

	"github.com/containerd/containerd"
	sshot "github.com/containerd/containerd/api/services/snapshots/v1"
)

var snapshotter = "overlayfs" // the default containerd uses -- https://github.com/containerd/containerd/blob/main/docs/cri/config.md#:~:text=snapshotter%20is%20the%20snapshotter%20used%20by%20containerd.

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	namespace := "t1"

	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		log.Fatal("client:", err)
	}
	defer client.Close()

	ssClient := sshot.NewSnapshotsClient(client.Conn())

	resp, err := ssClient.List(ctx, &sshot.ListSnapshotsRequest{Snapshotter: snapshotter})
	if err != nil {
		log.Fatal("snapshot list error:", err)
	}

	ssResp, err := resp.Recv()
	if err != nil {
		if strings.Contains(err.Error(), "EOF") {
			log.Println("all ok, no snapshots found")
			return
		}
		log.Fatal("snapshot recv error:", err)
	}

	ssMap := make(map[string]struct{})
	for _, ss := range ssResp.Info {
		log.Println(ss.Name, ss.Parent)
		
		_, ok := ssMap[ss.Name]
		if ok {
			delete(ssMap, ss.Name)
		} else {
			ssMap[ss.Name] = struct{}{}
		}

		// Ignore snapshots without parents, they are the ones we need to clean up.
		if ss.Parent == "" {
			continue
		}
		_, ok = ssMap[ss.Parent]
		if ok {
			delete(ssMap, ss.Parent)
		} else {
			ssMap[ss.Parent] = struct{}{}
		}
	}

	log.Println(ssMap)

	if len(ssMap) > 0 {
		var ssToRemove string
		for k := range ssMap {
			ssToRemove = k
		}

		resp2, err := ssClient.Remove(ctx, &sshot.RemoveSnapshotRequest{Snapshotter: snapshotter, Key: ssToRemove})
		if err != nil {
			log.Fatal("snapshot remove error:", err)
		}

		log.Println("removed", *resp2)
		return
	}
	log.Println("better luck next time snapshotter")
}