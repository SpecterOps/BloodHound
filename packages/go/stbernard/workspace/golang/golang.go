package golang

// GoPackage represents a parsed Go package
type GoPackage struct {
	Name   string `json:"name"`
	Dir    string `json:"dir"`
	Import string `json:"importpath"`
}
