package event

import (
	"os"
	"testing"

	"github.com/Meduzz/modulr/api"
	"github.com/gin-gonic/gin"
)

var (
	expectedData = ""
	protected    = false
	subject      = NewHttpDeliverer()
	srv          = gin.Default()
	service      = &api.DefaultService{
		ID:      "1",
		Name:    "test",
		Address: "localhost",
		Port:    6060,
		Context: "",
		Type:    "http",
		Scheme:  "http",
		Subscriptions: []*api.Subscription{
			{
				Topic:   "test",
				Routing: "test",
				Group:   "test",
				Path:    "/webhook",
				Secret:  "top secret",
			},
		},
	}
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
	err := subject.Deliver(service, service.GetSubscriptions()[0], []byte(text))

	if err != nil {
		t.Error(err)
	}
}

func TestUnhappyCase(t *testing.T) {
	text := "so unhapy!"
	expectedData = "something different"
	err := subject.Deliver(service, service.GetSubscriptions()[0], []byte(text))

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
	err := subject.Deliver(service, service.GetSubscriptions()[0], []byte(text))

	if err != nil {
		t.Error(err)
	}
}

func TestInvalidProtection(t *testing.T) {
	text := "so happy!"
	expectedData = text
	protected = true
	sub := service.GetSubscriptions()[0]
	sub.Secret = "asdf"

	err := subject.Deliver(service, sub, []byte(text))

	if err == nil {
		t.Error("expected an error")
	}

	if err.Error() != "call did not return 200" {
		t.Errorf("error message was not the expected one, was: %s", err.Error())
	}

	sub.Secret = "top secret"
}
