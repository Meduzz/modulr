package main

import (
	"log"

	"github.com/Meduzz/modulr"
	_ "github.com/Meduzz/modulr/adapter/event/adapter/nats"
	_ "github.com/Meduzz/modulr/adapter/event/delivery/http"
	_ "github.com/Meduzz/modulr/adapter/loadbalancer/roundrobin"
	_ "github.com/Meduzz/modulr/adapter/proxy/http"
	_ "github.com/Meduzz/modulr/adapter/registry/inmemory"
	"github.com/Meduzz/modulr/api"
	"github.com/gin-gonic/gin"
)

func main() {
	srv := gin.Default()

	// registers a service - naive version
	srv.POST("/register", func(ctx *gin.Context) {
		service := &api.DefaultService{
			Scheme: "http",
		}

		ctx.BindJSON(service)

		err := modulr.ServiceRegistry.Register(service)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		log.Printf("service named %s, of type %s, was registered\n", service.GetName(), service.GetType())

		ctx.Status(200)
	})

	// deregisters a service - naive version
	srv.DELETE("/deregister/:name/:id", func(ctx *gin.Context) {
		name := ctx.Param("name")
		id := ctx.Param("id")

		err := modulr.ServiceRegistry.Deregister(name, id)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Status(200)
	})

	srv.Any("/call/:service/*path", func(ctx *gin.Context) {
		name := ctx.Param("service")

		handler, err := modulr.HttpProxy.ForwarderFor(name)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		gin.WrapF(handler)(ctx)
	})

	srv.POST("/publish", func(ctx *gin.Context) {
		event := &api.Event{}
		ctx.BindJSON(event)

		err := modulr.EventSupport.Publish(event)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Status(200)
	})

	srv.Run(":8085")
}
