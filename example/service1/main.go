package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Meduzz/helper/http/client"
	"github.com/Meduzz/modulr/api"
	"github.com/gin-gonic/gin"
)

type (
	greetingLog struct {
		Name string `json:"name"`
	}

	test struct {
		*api.DefaultService
		Type string `json:"type"`
	}
)

func main() {
	port := flag.Int("port", 8086, "set port to start on")
	flag.Parse()

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

	register(*port)

	go deregister(*port)

	srv.Run(fmt.Sprintf(":%d", *port))
}

func register(port int) {
	subs := make([]*api.Subscription, 0)
	subs = append(subs, &api.Subscription{
		Topic: "service1.info",
		Path:  "/info",
		Group: "service1",
	})
	service := &api.DefaultService{
		ID:            fmt.Sprintf("%d", port),
		Name:          "service1",
		Address:       "localhost",
		Port:          port,
		Context:       "",
		Subscriptions: subs,
		Scheme:        "http",
	}
	test := &test{service, "test"}
	req, _ := client.POST("http://localhost:8085/register", test)
	req.Do(http.DefaultClient)
}

func deregister(id int) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	req, _ := client.DELETE(fmt.Sprintf("http://localhost:8085/deregister/service1/%d", id), nil)
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

	req, _ := client.POST("http://localhost:8085/publish", ev)
	req.Do(http.DefaultClient)
}
