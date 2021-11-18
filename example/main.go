package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	namespace := "t1"

	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		log.Fatal("client:", err)
	}
	defer client.Close()

	whaleCh := make(chan *Whale, 10)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(ctx context.Context, ch chan<- *Whale, wg *sync.WaitGroup){
		defer wg.Done()

		whale, err := newWhale(ctx, namespace, "docker.io/library/redis:alpine", "redis-server", client)
		if err != nil {
			whale.err = err
			ch <- whale

			fmt.Println("returning error:", err)
			return
		}
		err = whale.createContainer()
		whale.err = err
		ch <- whale

		fmt.Println("finished starting", whale.id)
		
	}(ctx, whaleCh, wg)

	fmt.Println("waiting.........")
	wg.Wait()
	fmt.Println("done waiting, now sleeping before closing the whale channel (imprison Willy!)")
	time.Sleep(2 * time.Second)
	close(whaleCh)

	var whales []*Whale
	for w := range whaleCh {
		whales = append(whales, w)

		fmt.Printf("\nid: %s\tpid: %d\n", w.id, w.pid)


	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	<-shutdown

	rdb := redis.NewClient(&redis.Options{
		Addr: "192.168.15.1:6379",
	})

	log.Println("waiting to ping")
	
	result, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Println("redis ping ", err)
	}
	log.Println(result)

	<-shutdown
	for _, w := range whales {
		if err := w.cleanUp(); err != nil {
			fmt.Println(err.Error())
		}
	}
	
	cancel()

	time.Sleep(2 * time.Second)
}

type Whale struct {
	ctx        context.Context
	client     *containerd.Client
	id         string
	namespace  string
	imageURL   string
	pid        uint32
	image      containerd.Image
	container  containerd.Container
	task       containerd.Task
	exitStatus <-chan containerd.ExitStatus
	err        error
}

func newWhale(ctx context.Context, namespace, imageURL, containerName string, client *containerd.Client) (*Whale, error) {
	if imageURL == "" {
		return nil, errors.New("image url cannot be empty")
	}
	if containerName == "" {
		return nil, errors.New("container name cannot be empty")
	}
	if client == nil {
		return nil, errors.New("containerd client cannot be nil")
	}

	fmt.Println("new whale comin: woot woot!")

	return &Whale{
		ctx: namespaces.WithNamespace(ctx, namespace), // overkill since containerd.WithDefaultNamespace() has been called, but oh well...
		client:    client,
		id:        containerName,
		namespace: namespace,
		imageURL:  imageURL,
	}, nil
}

func (w *Whale) createContainer() error {
	fmt.Println("creating container", w.id)
	
	var err error
	
	// pull the redis image from DockerHub
	w.image, err = w.client.Pull(w.ctx, w.imageURL, containerd.WithPullUnpack)
	if err != nil {
		return err
	}

	fmt.Println("done pulling image", w.imageURL)

	// create a container
	w.container, err = w.client.NewContainer(
		w.ctx,
		w.id,
		containerd.WithImage(w.image),
		containerd.WithNewSnapshot(w.id+"-snapshot", w.image),
		containerd.WithNewSpec(oci.WithImageConfig(w.image)),
	)
	if err != nil {
		return err
	}

	fmt.Println("creating task", w.id)

	// create a task from the container
	w.task, err = w.container.NewTask(w.ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return err
	}

	// make sure we wait before calling start
	w.exitStatus, err = w.task.Wait(w.ctx)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("startin", w.id)
	// call start on the task to execute the redis server
	if err := w.task.Start(w.ctx); err != nil {
		return err
	}

	w.pid = w.task.Pid()

	return nil
}

func (w *Whale) cleanUp() error {
	defer func(){
		if w.task != nil {
			w.task.Delete(w.ctx)
		}
		if w.container != nil {
			w.container.Delete(w.ctx, containerd.WithSnapshotCleanup)
		}
	}()

	if w.task != nil {
		if err := w.task.Kill(w.ctx, syscall.SIGTERM); err != nil {
			return err
		}
		status := <-w.exitStatus
		code, _, err := status.Result()
		if err != nil {
			return err
		}
	
		fmt.Printf("\n%s killed with status: %d\n", w.id, code)
		return nil
	}
	fmt.Println("\n", "nothing to see here... ", w.id)

	return errors.Errorf("nothin happened :(( %s", w.id)
}
