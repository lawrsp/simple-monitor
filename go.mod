module coding.pickflames.com/monitor

go 1.16

replace coding.pickflames.com/pickflames/framework => ../framework

require (
	coding.pickflames.com/pickflames/framework v0.0.0-00010101000000-000000000000
	github.com/containerd/containerd v1.5.0 // indirect
	github.com/docker/docker v20.10.6+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/gin-gonic/gin v1.7.7
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/urfave/cli v1.22.5
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/grpc v1.37.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
