package static

import (
	"embed"

	"github.com/specterops/bloodhound/src/api"
)

const (
	assetBasePath  = "assets"
	indexAssetPath = "index.html"
)

//go:embed all:assets
var assets embed.FS

var AssetHandler = MakeAssetHandler(AssetConfig{
	FS:         assets,
	BasePath:   assetBasePath,
	IndexPath:  indexAssetPath,
	PrefixPath: api.UserInterfacePath,
})
