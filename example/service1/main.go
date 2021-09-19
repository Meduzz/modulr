package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Meduzz/helper/http/client"
	"github.com/Meduzz/modulr/api"
	"github.com/gin-gonic/gin"
)

type greetingLog struct {
	Name string `json:"name"`
}

func main() {
	srv := gin.Default()

	srv.GET("/hello/:who", func(ctx *gin.Context) {
		who := ctx.Param("who")

		ctx.String(200, "Hello %s!", who)

		go sendEvent(who)
	})

	srv.POST("/info", func(ctx *gin.Context) {
		info := &greetingLog{}
		ctx.BindJSON(info)

		log.Printf("Greeted %s\n", info.Name)
	})

	register()

	go deregister()

	srv.Run(":8081")
}

func register() {
	subs := make([]*api.Subscription, 0)
	subs = append(subs, &api.Subscription{
		Topic: "service1.info",
		Path:  "/info",
		Group: "service1",
	})
	service := api.Service{
		ID:            "service1",
		Name:          "service1",
		Address:       "localhost",
		Port:          8081,
		Context:       "",
		Subscriptions: subs,
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

func sendEvent(greeted string) {
	greeting := &greetingLog{
		Name: greeted,
	}

	bs, _ := json.Marshal(greeting)

	ev := api.Event{
		Topic: "service1.info",
		Body:  json.RawMessage(bs),
	}

	req, _ := client.POST("http://localhost:8080/publish", ev)
	req.Do(http.DefaultClient)
}
