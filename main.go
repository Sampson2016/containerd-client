package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func main() {
	client, err := containerd.New(`\\.\pipe\containerd-containerd`)
	defer client.Close()
	if err != nil {
		panic(err)
	}

	ctx := namespaces.WithNamespace(context.Background(), "default")

	bundleSpec := &specs.Spec{}

	container, err := client.NewContainer(ctx, "test2-winc", containerd.WithRuntime("io.containerd.runtime.v1.winc", nil), containerd.WithSpec(bundleSpec))
	if err != nil {
		panic(err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStreams(os.Stdin, os.Stdout, os.Stderr)))
	if err != nil {
		panic(err)
	}

	fmt.Printf("task: %+v\n", task)

	containers, err := client.Containers(ctx)
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		fmt.Printf("container: %+v\n", c)
	}
}
