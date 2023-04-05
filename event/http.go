package event

import (
	"fmt"
	"net/http"

	"github.com/Meduzz/helper/http/client"
	"github.com/Meduzz/modulr/api"
)

type httpAdapter struct{}

func NewHttpDeliverer() EventDeliveryAdapter {
	return &httpAdapter{}
}

func (h *httpAdapter) Deliver(service api.Service, sub *api.Subscription, body []byte) error {
	url := ""

	if service.GetPort() != 0 {
		url = fmt.Sprintf("%s://%s:%d%s%s", service.GetScheme(), service.GetAddress(), service.GetPort(), service.GetContext(), sub.Path)
	} else {
		url = fmt.Sprintf("%s://%s%s%s", service.GetScheme(), service.GetAddress(), service.GetContext(), sub.Path)
	}

	req, err := client.POSTBytes(url, body, "application/json")

	if err != nil {
		return err
	}

	if sub.Secret != "" {
		req.Header("Authorization", sub.Secret)
	}

	res, err := req.Do(http.DefaultClient)

	if err != nil {
		return err
	}

	if res.Code() != 200 {
		return fmt.Errorf("call did not return 200")
	}

	return nil
}
