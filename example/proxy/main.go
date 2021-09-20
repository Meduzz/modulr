package main

import (
	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/modulr/api"
	deliverer "github.com/Meduzz/modulr/delivery/adapter"
	"github.com/Meduzz/modulr/event"
	eventadapter "github.com/Meduzz/modulr/event/adapter"
	proxyadapter "github.com/Meduzz/modulr/proxy/adapter"
	"github.com/Meduzz/modulr/registry"
	"github.com/gin-gonic/gin"
)

func main() {
	srv := gin.Default()

	conn, err := nuts.Connect()

	if err != nil {
		panic(err)
	}

	eventing := eventadapter.NewNatsAdapter(conn)
	deliveryadapter := deliverer.NewHttpAdapter()

	loadbalancer := proxyadapter.NewLoadBalancer()
	eventhandler := event.NewEventRegistry(eventing, deliveryadapter)

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
