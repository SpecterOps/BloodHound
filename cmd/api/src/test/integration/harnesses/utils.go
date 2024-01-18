package harnesses

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func Filename() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("unable to get the current filename")
	}
	return filename, nil
}

func Dirname() (string, error) {
	filename, err := Filename()
	if err != nil {
		return "", err
	}
	return filepath.Dir(filename), nil
}

type NodeStyle struct {
	NodeColor string `json:"node-color"`
}

type RelationshipStyle struct {
	NodeColor string `json:"arrow-color"`
}

type Node struct {
	ID         string            `json:"id"`
	Caption    string            `json:"caption"`
	Labels     []string          `json:"labels"`
	Properties map[string]string `json:"properties"`
	Style      NodeStyle
}

type Relationship struct {
	ID         string            `json:"id"`
	From       string            `json:"fromId"`
	To         string            `json:"toId"`
	Kind       string            `json:"type"`
	Properties map[string]string `json:"properties"`
	Style      RelationshipStyle
}

type HarnessData struct {
	Nodes         []Node
	Relationships []Relationship
}

func ReadHarness(harnessName string) (HarnessData, error) {
	if dir, err := Dirname(); err != nil {
		return HarnessData{}, err
	} else if jsonFile, err := os.Open(path.Join(dir, harnessName+".json")); err != nil {
		return HarnessData{}, err
	} else {
		defer jsonFile.Close()

		if byteValue, err := io.ReadAll(jsonFile); err != nil {
			return HarnessData{}, err
		} else {
			var result HarnessData

			json.Unmarshal([]byte(byteValue), &result)

			return result, nil
		}
	}
}
