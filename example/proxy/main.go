package main

import (
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/event"
	"github.com/Meduzz/modulr/event/natsadapter"
	"github.com/Meduzz/modulr/proxy/httpadapter"
	"github.com/Meduzz/modulr/registry"
	"github.com/gin-gonic/gin"
)

func main() {
	srv := gin.Default()

	loadbalancer := httpadapter.NewLoadBalancer()
	eventadapter, _ := natsadapter.NewNatsAdapter()
	eventhandler := event.NewEventRegistry(eventadapter)
	serviceRegistry := registry.NewServiceRegistry(loadbalancer, eventhandler)

	// registers a service - naive version
	srv.POST("/register", func(ctx *gin.Context) {
		service := &api.Service{}
		ctx.BindJSON(service)

		err := serviceRegistry.Register(service)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Status(200)
	})

	// deregisters a service - naive version
	srv.DELETE("/deregister/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")

		err := serviceRegistry.Deregister(id)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Status(200)
	})

	srv.Any("/call/:service/*path", func(ctx *gin.Context) {
		name := ctx.Param("service")

		handler := loadbalancer.Lookup(name)

		if handler == nil {
			ctx.Status(404)
			return
		}

		delegate := gin.WrapF(handler)
		delegate(ctx)
	})

	srv.POST("/publish", func(ctx *gin.Context) {
		event := &api.Event{}
		ctx.BindJSON(event)

		err := eventhandler.Publish(event)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Status(200)
	})

	srv.Run(":8080")
}
