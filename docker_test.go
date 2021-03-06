package main

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	fmt.Println("Test started")

	os.Exit(m.Run())
}

func TestClient(t *testing.T) {

	c, err := NewDockerClient(&DockerConfig{})
	if err != nil {
		t.Errorf("new client failed: %v", err)
	}
	ctx := context.Background()

	list, err := c.ListContainers(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(list)

	param := &UpdateServiceParam{Image: "hello-world:latest"}

	if err := c.UpdateService(ctx, param, os.Stdout); err != nil {
		fmt.Println(err)
	}

}
