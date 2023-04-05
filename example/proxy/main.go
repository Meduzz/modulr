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

type (
	// Test extends the api.Service (As an example)
	Test interface {
		api.Service
		GetType() string
	}

	test struct {
		*api.DefaultService
		Type string `json:"type"`
	}
)

func main() {
	srv := gin.Default()

	conn, err := nuts.Connect()

	if err != nil {
		panic(err)
	}

	factory := loadbalancer.NewRoundRobinFactory()
	eventing := event.NewNatsAdapter(conn)
	deliveryadapter := event.NewHttpDeliveryAdapter()

	serviceRegistry := registry.NewServiceRegistry()

	loadbalancer := proxy.NewHttpProxy(factory)

	eventSupport := event.NewEventSupport(serviceRegistry, eventing, factory)
	eventSupport.RegisterDeliveryAdapter("http", deliveryadapter)

	// registers a service - naive version
	srv.POST("/register", func(ctx *gin.Context) {
		service := &test{
			DefaultService: &api.DefaultService{
				Scheme: "http",
			},
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

		services := serviceRegistry.Lookup(name)
		lb := factory.For(name)
		service := lb.Next(services)

		test, ok := service.(Test)

		if !ok {
			log.Println("could not cast service to Test")
		} else {
			log.Printf("calling %s of type %s\n", test.GetName(), test.GetType())
		}

		handler := loadbalancer.ForwarderFor(service)

		if handler == nil {
			ctx.Status(404)
			return
		}

		handler(ctx)
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

func (t *test) GetType() string {
	return t.Type
}
