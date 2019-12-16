// Simple client for submitting commands over HTTP
package httpcommandclient

import (
	"context"
	"fmt"
	"github.com/function61/eventkit/command"
	"github.com/function61/eventkit/httpcommand"
	"github.com/function61/gokit/ezhttp"
	"net/http"
)

type Client struct {
	baseUrl     string // looks like "http://localhost/command/"
	bearerToken string
	httpClient  *http.Client
}

func New(baseUrl string, bearerToken string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseUrl:     baseUrl,
		bearerToken: bearerToken,
		httpClient:  httpClient,
	}
}

func (c *Client) Exec(ctx context.Context, cmdStruct command.Command) error {
	_, err := c.execInternal(ctx, cmdStruct)
	return err
}

func (c *Client) ExecExpectingCreatedRecordId(ctx context.Context, cmdStruct command.Command) (string, error) {
	res, err := c.execInternal(ctx, cmdStruct)
	if err != nil {
		return "", err
	}

	collectionId := res.Header.Get(httpcommand.CreatedRecordIdHeaderKey)
	if collectionId == "" {
		return "", fmt.Errorf("didn't get back a %s header", httpcommand.CreatedRecordIdHeaderKey)
	}

	return collectionId, nil
}

func (c *Client) execInternal(ctx context.Context, cmdStruct command.Command) (*http.Response, error) {
	if err := cmdStruct.Validate(); err != nil {
		return nil, err
	}

	res, err := ezhttp.Post(
		ctx,
		c.baseUrl+cmdStruct.Key(),
		ezhttp.AuthBearer(c.bearerToken),
		ezhttp.SendJson(cmdStruct),
		ezhttp.Client(c.httpClient))
	if err != nil {
		return nil, err
	}

	return res, nil
}
