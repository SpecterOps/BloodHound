package os

import "os"

//go:generate go run go.uber.org/mock/mockgen -destination=./mocks/mock.go -package=mocks . OSService
type Service interface {
	Create(name string) (*os.File, error)
	CreateTemporaryDirectory(dir, pattern string) (*os.File, error)
	MakeDirectory(name string, perm os.FileMode) error
	ReadFile(name string) ([]byte, error)
	Remove(path string) error
}

type Client struct {
	Service Service
}

func New() *Client {
	return &Client{}
}

func (c *Client) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (c *Client) CreateTemporaryDirectory(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func (c *Client) MakeDirectory(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (c *Client) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (c *Client) Remove(name string) error {
	return os.Remove(name)
}
