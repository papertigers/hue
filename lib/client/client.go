package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/papertigers/hue/lib/bridge"
)

type Client struct {
	client   *http.Client
	username string
	baseurl  string
}

func NewClient(username string, bridge bridge.Bridge) *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		username: username,
		baseurl:  fmt.Sprintf("http://%s/api/", bridge.IP),
	}
}

/*
 * A lot of this is taken from https://github.com/joyent/triton-go/
 */

type RequestInput struct {
	Method string
	Path   string
	Body   interface{}
}

func (c *Client) ExecuteRequest(ctx context.Context, input RequestInput) (io.ReadCloser, error) {

	method := input.Method
	path := input.Path
	body := input.Body

	var requestBody io.ReadSeeker
	if body != nil {
		marshaled, err := json.MarshalIndent(body, "", "    ")
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewReader(marshaled)
	}

	endpoint := fmt.Sprintf("%s/%s/%s", c.baseurl, c.username, path)
	req, err := http.NewRequest(method, endpoint, requestBody)
	if err != nil {
		return nil, errwrap.Wrapf("Error constructing HTTP requst: {{err}}", err)
	}

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errwrap.Wrapf("Error executing HTTP request: {{err}}", err)
	}

	/*
	 * Hue always returns a 200 unless something is wrong.
	 * The response generally has json alerting you if the req was a success
	 * or failure.
	 */
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP bad status code: %s", resp.StatusCode)
	}

	return resp.Body, nil
}
