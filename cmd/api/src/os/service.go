package os

import "os"

//go:generate go run go.uber.org/mock/mockgen -destination=./mocks/mock.go -package=mocks . OSService
type OSService interface {
	CreateTemporaryDirectory(dir, pattern string) (*os.File, error)
}

type Client struct {
	Service OSService
}

func New() *Client {
	return &Client{}
}

func (c *Client) CreateTemporaryDirectory(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}
