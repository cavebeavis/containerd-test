package main

import (
	"context"
	"log"

	"github.com/containerd/containerd"
	imgs "github.com/containerd/containerd/api/services/images/v1"
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

	imgsClient := imgs.NewImagesClient(client.Conn())

	resp, err := imgsClient.List(ctx, &imgs.ListImagesRequest{})
	if err != nil {
		log.Fatal("images response error:", err)
	}

	if len(resp.Images) > 0 {
		log.Println(resp.Images)

		for _, img := range resp.Images {
			// Ignore the response which is just *types.Empty -- does us no good at this point...
			_, err := imgsClient.Delete(ctx, &imgs.DeleteImageRequest{Name: img.Name})
			if err != nil {
				log.Fatal("image delete error:", err)
			}
			log.Println("successfully delete image: ", img)
		}
		return
	}
	
	log.Println("no images to delete... good for you!")
}