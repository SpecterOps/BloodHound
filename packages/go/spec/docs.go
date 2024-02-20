package spec

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed v2
var openApiSrc embed.FS

var OpenApiSpecHandler http.Handler

func init() {
	if fSys, err := fs.Sub(openApiSrc, "v2"); err != nil {
		panic(err)
	} else {
		OpenApiSpecHandler = http.FileServer(http.FS(fSys))
	}
}
