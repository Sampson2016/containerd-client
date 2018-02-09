package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const rootfs = "C:\\ProgramData\\docker\\windowsfilter\\af4b828a79120f7c72f60f10f1b359fcd90ff70432cd1536b155a6814526b69e"

func main() {
	client, err := containerd.New(`\\.\pipe\containerd-containerd`)
	defer client.Close()
	if err != nil {
		panic(err)
	}

	containerId := os.Args[1]

	ctx := namespaces.WithNamespace(context.Background(), "default")

	bundleSpec, err := generateSpec(containerId)
	if err != nil {
		panic(err)
	}
	bundleSpec.Process = &specs.Process{
		Cwd:  "C:\\",
		Args: []string{"cmd.exe"},
	}

	container, err := client.NewContainer(ctx, containerId, containerd.WithRuntime("io.containerd.runtime.v1.winc", nil), containerd.WithSpec(bundleSpec))
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

func generateSpec(containerId string) (*specs.Spec, error) {
	wincImageBinary := `c:\workspace\winc-image-binary\winc-image.exe`
	exec.Command(wincImageBinary, "delete", containerId).Run()

	o, err := exec.Command(wincImageBinary, "create", rootfs, containerId).CombinedOutput()
	if err != nil {
		fmt.Println(string(o))
		return nil, err
	}

	var b specs.Spec
	if err := json.Unmarshal(o, &b); err != nil {
		return nil, err
	}

	return &b, nil
}
