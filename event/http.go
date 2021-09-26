package event

import (
	"fmt"
	"net/http"

	"github.com/Meduzz/helper/http/client"
)

type httpAdapter struct{}

func NewHttpDeliveryAdapter() DeliveryAdapter {
	return &httpAdapter{}
}

func (h *httpAdapter) DeliverEvent(url string, body []byte) error {
	req, err := client.POSTBytes(url, body, "application/json")

	if err != nil {
		return err
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
