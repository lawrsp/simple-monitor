package main

import (
	"context"
	"fmt"
	"net/http"

	"coding.pickflames.com/pickflames/framework/rest"
	"coding.pickflames.com/pickflames/framework/utils/log"
	"github.com/gin-gonic/gin"
)

type RESTConfig struct {
	Address     string
	AccessToken string
	UseH2C      bool
}

type Server struct {
	Config *RESTConfig
	Engine *gin.Engine
}

func NewServer(c *RESTConfig) (*Server, error) {
	if c.Address == "" {
		return nil, fmt.Errorf("No Adderss")
	}
	return &Server{Config: c}, nil
}

func (s *Server) InitRoutes(dc *DockerClient) error {

	engine := gin.New(func(e *gin.Engine) {
		if s.Config.UseH2C {
			e.UseH2C = true
		}
	})

	engine.Use(gin.Logger())
	engine.Use(gin.HandlerFunc(rest.CheckAndRecovery()))
	// check permission
	engine.Use(func(c *gin.Context) {
		token := c.Query("access_token")
		if token != s.Config.AccessToken {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	})

	engine.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	service := engine.Group("/service")
	service.GET("/list", func(c *gin.Context) {
		// docker service ls
		ctx := context.Background()

		list, err := dc.ListServices(ctx)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, list)
	})

	service.POST("/update", func(c *gin.Context) {
		// docker service update ...
		ctx := context.Background()
		param := &UpdateServiceParam{}
		if err := c.ShouldBindJSON(param); err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Header("Transfer-Encoding", "chunked")
		err := dc.UpdateService(ctx, param, c.Writer)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.AbortWithStatus(http.StatusOK)
		// c.JSON(http.StatusOK, gin.H{"result": "OK"})
	})

	s.Engine = engine

	return nil
}

func (s *Server) Run() error {
	bind := s.Config.Address
	if bind == "" {
		return fmt.Errorf("no address")
	}

	log.Infof("Service(%s) start.", bind)
	err := s.Engine.Run(bind)
	if err != nil {
		log.Errorf("Service(%s) failed: %v.", bind, err)
		return err
	}

	return nil
}
