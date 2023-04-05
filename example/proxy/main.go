package main

import (
	"log"

	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/event"
	"github.com/Meduzz/modulr/loadbalancer"
	"github.com/Meduzz/modulr/proxy"
	"github.com/Meduzz/modulr/registry"
	"github.com/gin-gonic/gin"
)

func main() {
	srv := gin.Default()

	conn, err := nuts.Connect()

	if err != nil {
		panic(err)
	}

	factory := loadbalancer.NewRoundRobinFactory()
	eventing := event.NewNatsAdapter(conn)
	deliveryadapter := event.NewHttpDeliverer()

	serviceRegistry := registry.NewServiceRegistry()
	inmemStorage := registry.NewInMemoryStorage()
	serviceRegistry.SetStorage(inmemStorage)

	httpProxy := proxy.NewProxy()
	httpForwarder := proxy.NewHttpForwarder()
	httpProxy.RegisterForwarder("http", httpForwarder)

	eventSupport := event.NewEventSupport(serviceRegistry, eventing, factory)
	eventSupport.RegisterDeliverer("http", deliveryadapter)

	// registers a service - naive version
	srv.POST("/register", func(ctx *gin.Context) {
		service := &api.DefaultService{
			Scheme: "http",
		}

		ctx.BindJSON(service)

		err := serviceRegistry.Register(service)

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

		err := serviceRegistry.Deregister(name, id)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Status(200)
	})

	srv.Any("/call/:service/*path", func(ctx *gin.Context) {
		name := ctx.Param("service")

		services, err := serviceRegistry.Lookup(name)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		if len(services) == 0 {
			ctx.Status(404)
			return
		}

		lb := factory.For(name)
		service := lb.Next(services)

		handler := httpProxy.ForwarderFor(service)

		gin.WrapF(handler)(ctx)
	})

	srv.POST("/publish", func(ctx *gin.Context) {
		event := &api.Event{}
		ctx.BindJSON(event)

		err := eventing.Publish(event.Topic, event.Routing, event.Body)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Status(200)
	})

	srv.Run(":8085")
}
