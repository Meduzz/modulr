package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/Meduzz/helper/http/client"
	"github.com/Meduzz/modulr/api"
	"github.com/gin-gonic/gin"
)

func main() {
	srv := gin.Default()

	srv.GET("/hello/:who", func(ctx *gin.Context) {
		who := ctx.Param("who")

		ctx.String(200, "Hello %s!", who)
	})

	register()

	go deregister()

	srv.Run(":8081")
}

func register() {
	service := api.Service{
		ID:      "service1",
		Name:    "service1",
		Address: "localhost",
		Port:    8081,
		Context: "",
	}
	req, _ := client.POST("http://localhost:8080/register", service)
	req.Do(http.DefaultClient)
}

func deregister() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	req, _ := client.DELETE("http://localhost:8080/deregister/service1", nil)
	req.Do(http.DefaultClient)

	os.Exit(0)
}
