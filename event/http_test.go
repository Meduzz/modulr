package event

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

var (
	expectedData = ""
	protected    = false
	subject      = NewHttpDeliveryAdapter()
	srv          = gin.Default()
)

func start() {
	srv.Run(":6060")
}

func TestMain(m *testing.M) {
	srv.POST("/webhook", func(ctx *gin.Context) {
		if protected {
			token := ctx.GetHeader("Authorization")

			if token == "" || token != "top secret" {
				ctx.AbortWithStatus(400)
				return
			}
		}

		bs, err := ctx.GetRawData()

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		if expectedData != string(bs) {
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
	expectedData = text
	err := subject.DeliverEvent("http://localhost:6060/webhook", "", []byte(text))

	if err != nil {
		t.Error(err)
	}
}

func TestUnhappyCase(t *testing.T) {
	text := "so unhapy!"
	expectedData = "something different"
	err := subject.DeliverEvent("http://localhost:6060/webhook", "", []byte(text))

	if err == nil {
		t.Errorf("expected an error")
	}

	if err.Error() != "call did not return 200" {
		t.Errorf("error message was not the expected one, was: %s", err.Error())
	}
}

func TestHappyProtectionOn(t *testing.T) {
	text := "so happy!"
	expectedData = text
	protected = true
	err := subject.DeliverEvent("http://localhost:6060/webhook", "top secret", []byte(text))

	if err != nil {
		t.Error(err)
	}
}

func TestInvalidProtection(t *testing.T) {
	text := "so happy!"
	expectedData = text
	protected = true
	err := subject.DeliverEvent("http://localhost:6060/webhook", "", []byte(text))

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "call did not return 200" {
		t.Errorf("error message was not the expected one, was: %s", err.Error())
	}

	err = subject.DeliverEvent("http://localhost:6060/webhook", "asdf", []byte(text))

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "call did not return 200" {
		t.Errorf("error message was not the expected one, was: %s", err.Error())
	}
}
