package main

import (
	"github.com/Meduzz/modulr/api"
	"github.com/Meduzz/modulr/proxy/httpadapter"
	"github.com/Meduzz/modulr/registry"
	"github.com/gin-gonic/gin"
)

func main() {
	srv := gin.Default()

	loadbalancer := httpadapter.NewLoadBalancer()
	serviceRegistry := registry.NewServiceRegistry(loadbalancer)

	// registers a service - naive version
	srv.POST("/register", func(ctx *gin.Context) {
		service := &api.Service{}
		ctx.BindJSON(service)

		// TODO what can go wrong here?
		serviceRegistry.Register(service)

		ctx.Status(200)
	})

	// deregisters a service - naive version
	srv.DELETE("/deregister/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")

		// TODO what can go wrong here?
		serviceRegistry.Deregister(id)

		ctx.Status(200)
	})

	srv.Any("/call/:service/*path", func(ctx *gin.Context) {
		name := ctx.Param("service")

		handler := loadbalancer.Lookup(name)

		if handler == nil {
			ctx.Status(404)
		}

		delegate := gin.WrapF(handler)
		delegate(ctx) // is this enough?
	})

	srv.POST("/publish")

	srv.Run(":8080")
}
