package main

import (
	"context"
	"log"

	"github.com/containerd/containerd"
	nspcs "github.com/containerd/containerd/api/services/namespaces/v1"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	namespace := "t1"

	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		log.Fatal("client:", err)
	}
	defer client.Close()

	nspcClient := nspcs.NewNamespacesClient(client.Conn())

	/*resp, err := nspcClient.Create(ctx, &nspcs.CreateNamespaceRequest{Namespace: nspcs.Namespace{Name: namespace}})
	if err != nil {
		log.Fatal("namespace create error:", err)
	}

	log.Println("create:", *resp)*/

	/*resp2, err := nspcClient.Delete(ctx, &nspcs.DeleteNamespaceRequest{Name: namespace})
	if err != nil {
		log.Fatal("namespace delete error:", err)
	}*/

	

	list, err := nspcClient.List(ctx, &nspcs.ListNamespacesRequest{/*Filter: namespace*/})
	if err != nil {
		log.Fatal("namespace list error:", err)
	}

	if len(list.Namespaces) > 0 {
		log.Println("list:", list.Namespaces)

		for _, ns := range list.Namespaces {
			if ns.Name == namespace {
				delResp, err := nspcClient.Delete(ctx, &nspcs.DeleteNamespaceRequest{Name: namespace})
				if err != nil {
					log.Fatal("namespace delete error:", err)
				}

				log.Println("delete resp:", *delResp)
			}
		}
		return
	}
	
}