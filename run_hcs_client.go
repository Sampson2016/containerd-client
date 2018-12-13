package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/namespaces"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func main() {
	client, err := containerd.New(`\\.\pipe\containerd-containerd`)
	defer client.Close()
	if err != nil {
		panic(err)
	}

	containerId := os.Args[1]

	bundleSpec, err := getBundleSpec(containerId)
	if err != nil {
		panic(err)
	}

	bundleSpec.Process = &specs.Process{
		Cwd:  "C:\\",
		Args: []string{"powershell.exe", "-Command", "echo 'hi from a runhcs container'"},
	}
	bundleSpec.Mounts = []specs.Mount{
		specs.Mount{
			Destination: "C:\\root",
			Source:      "C:\\workspace\\go",
		},
	}

	// we can't fill in the rootfs and the WindowsLayerPaths in the spec we give to the container
	// instead, we pass that information in when we create the task
	if bundleSpec.Root != nil {
		panic("rootfs not nil")
	}

	windowsLayers := bundleSpec.Windows.LayerFolders
	bundleSpec.Windows.LayerFolders = nil

	ctx := namespaces.WithNamespace(context.Background(), "default")
	container, err := client.NewContainer(ctx,
		containerId,
		containerd.WithRuntime("io.containerd.runhcs.v1", nil),
		containerd.WithSpec(&bundleSpec),
	)

	parentLayers, err := json.Marshal(windowsLayers[:len(windowsLayers)-1])
	if err != nil {
		panic(err)
	}

	rootfs := mount.Mount{
		Type:    "windows-layer",
		Source:  windowsLayers[len(windowsLayers)-1],                                             // scratch layer
		Options: []string{fmt.Sprintf("%s%s", mount.ParentLayerPathsFlag, string(parentLayers))}, // parent layers
	}

	task, err := container.NewTask(ctx,
		cio.NewCreator(cio.WithStreams(os.Stdin, os.Stdout, os.Stderr)),
		containerd.WithRootFS([]mount.Mount{rootfs}))
	if err != nil {
		panic(err)
	}

	if err := task.Start(ctx); err != nil {
		panic(err)
	}

	status, err := task.Wait(ctx)
	if err != nil {
		panic(err)
	}

	result := <-status
	fmt.Printf("result: %+v\n", result)

	if _, err := task.Delete(ctx); err != nil {
		panic(err)
	}

	if err := container.Delete(ctx); err != nil {
		panic(err)
	}
}

func getBundleSpec(containerId string) (specs.Spec, error) {
	grootBin := `c:\workspace\go\bin\groot-windows.exe`
	exec.Command(grootBin, "--driver-store", "c:\\programdata\\groot", "delete", containerId).Run()

	o, err := exec.Command(grootBin, "--driver-store", "c:\\programdata\\groot", "create", "docker:///cloudfoundry/windows2016fs:1803", containerId).Output()
	if err != nil {
		fmt.Println(string(o))
		return specs.Spec{}, err
	}

	var b specs.Spec
	if err := json.Unmarshal(o, &b); err != nil {
		return specs.Spec{}, err
	}

	return b, nil
}
