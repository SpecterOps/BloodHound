package os

import "os"

//go:generate go run go.uber.org/mock/mockgen -destination=./mocks/mock.go -package=mocks . Service
type Service interface {
	CreateTemporaryDirectory(dir, pattern string) (*os.File, error)
}

type Client struct {}

func (c *Client) CreateTemporaryDirectory(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}
