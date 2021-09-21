package adapter

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

var (
	event   = ""
	subject = NewHttpAdapter()
	srv     = gin.Default()
)

func start() {
	srv.Run(":8080")
}

func TestMain(m *testing.M) {
	srv.POST("/webhook", func(ctx *gin.Context) {
		bs, err := ctx.GetRawData()

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		if event != string(bs) {
			ctx.Status(400)
			return
		}

		ctx.Status(200)
	})

	go start()

	os.Exit(m.Run())
}

func TestHappyCase(t *testing.T) {
	text := "so happy!"
	event = text
	err := subject.DeliverEvent("http://localhost:8080/webhook", []byte(text))

	if err != nil {
		t.Error(err)
	}
}

func TestUnhappyCase(t *testing.T) {
	text := "so unhapy!"
	event = "something different"
	err := subject.DeliverEvent("http://localhost:8080/webhook", []byte(text))

	if err == nil {
		t.Errorf("expected an error")
	}

	if err.Error() != "call did not return 200" {
		t.Errorf("error message was not the expected one, was: %s", err.Error())
	}
}
