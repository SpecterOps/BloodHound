package handlers

type ExtensionView struct {
	ExtensionID int32  `json:"extension_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Namespace   string `json:"namespace"`
	Version     string `json:"version"`
}
