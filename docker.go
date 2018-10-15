package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	yaml "gopkg.in/yaml.v2"

	"coding.pickflames.com/pickflames/framework/utils/fileutil"
	"coding.pickflames.com/pickflames/framework/utils/log"
)

type DockerConfig struct {
	RegistryFile string
}

type DockerClient struct {
	*client.Client
	auths map[string]*types.AuthConfig
}

func NewDockerClient(conf *DockerConfig) (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	cli.NegotiateAPIVersion(context.Background())

	c := &DockerClient{Client: cli, auths: map[string]*types.AuthConfig{}}

	// read registry config
	if err := c.readRegistryAuths(conf.RegistryFile); err != nil {
		log.Errorf("read registry auth config failed: %v", err)
		return nil, err
	}

	return c, nil
}

func (c *DockerClient) readRegistryAuths(filePath string) error {
	if filePath == "" {
		return nil
	}

	bs, err := fileutil.LoadDataFrom(filePath)
	if err != nil {
		log.Errorf("load data from %s failed: %v", filePath, err)
		return err
	}

	conf := map[string]*types.AuthConfig{}
	err = yaml.Unmarshal(bs, conf)
	if err != nil {
		log.Errorf("parse registry file %s failed: %v", filePath, err)
		return err
	}

	for k, v := range conf {
		c.auths[k] = v
	}

	return nil
}

func (c *DockerClient) ListContainers(ctx context.Context) ([]string, error) {

	containers, err := c.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}

	return nil, nil
}

func (c *DockerClient) ListServices(ctx context.Context) (interface{}, error) {

	services, err := c.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, err
	}

	return services, nil
}

type QueryServiceParam struct {
	Image string `json:"image"`
}

type UpdateServiceParam struct {
	Image string `json:"image"`
}

func (p *UpdateServiceParam) Match(dsvc *swarm.Service) bool {
	//image: xxxxxx/xxxxx

	image := dsvc.Spec.TaskTemplate.ContainerSpec.Image
	qimg := p.Image

	imgOld := strings.Split(image, "@")[0]
	imgNew := strings.Split(qimg, "@")[0]
	if imgOld != imgNew {
		return false
	}

	verOld := strings.Split(imgOld, ":")[1]
	verNew := strings.Split(imgNew, ":")[1]

	if !strings.HasPrefix(verOld, verNew) {
		return false
	}

	return true
}

func (c *DockerClient) UpdateService(ctx context.Context, param *UpdateServiceParam) error {
	if param.Image == "" {
		return fmt.Errorf("no image")
	}

	var authStr string
	if names := strings.Split(param.Image, "/"); len(names) > 1 {
		url := names[0]
		authConfig := c.auths[url]
		if authConfig != nil {
			encodedJSON, err := json.Marshal(authConfig)
			if err != nil {
				panic(err)
			}
			authStr = base64.URLEncoding.EncodeToString(encodedJSON)
		}
	}

	out, err := c.ImagePull(ctx, param.Image, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return fmt.Errorf("pull image failed: %v", err)
	}

	defer out.Close()

	io.Copy(os.Stdout, out)

	// find the image used by which service
	services, err := c.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return err
	}

	for _, dsvc := range services {

		if param.Match(&dsvc) {
			newSpec := &swarm.ServiceSpec{}
			*newSpec = dsvc.Spec
			newContainerSpec := &swarm.ContainerSpec{}
			*newContainerSpec = *dsvc.Spec.TaskTemplate.ContainerSpec
			newContainerSpec.Image = param.Image
			newSpec.TaskTemplate.ContainerSpec = newContainerSpec

			result, err := c.ServiceUpdate(ctx, dsvc.ID, dsvc.Version, *newSpec, types.ServiceUpdateOptions{
				QueryRegistry:       true,
				EncodedRegistryAuth: authStr,
				// RegistryAuthFrom:    types.RegistryAuthFromPreviousSpec,
			})
			if err != nil {
				log.Errorf("updat service %s failed: %v", dsvc.Spec.Name)
				return err
			}

			log.Infof("service %s updated: %v", dsvc.Spec.Name, result)
		}
	}

	return nil

}
