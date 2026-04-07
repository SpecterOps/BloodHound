// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package bloodhoundgraph

import (
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

//TODO: Move styling responsibilities to the UI or move shared styling definitions to a cue file to generate from one source of truth

var (
	fontAwesomePrefix  = "fas fa-"
	defaultUnknownIcon = "fas fa-question"
	defaultIconColor   = "#EEE"
	metaNodeKindLabel  = "Meta"
	metaNodeColor      = "#000"
	metaImageDefault   = "/ui/meta.png"
	metaImageTierZero  = "/ui/metat0.png"
)

type BloodHoundGraphGlyph struct {
	Angle    int                        `json:"angle,omitempty"`
	Blink    bool                       `json:"blink,omitempty"`
	Border   *BloodHoundGraphItemBorder `json:"border,omitempty"`
	Color    string                     `json:"color,omitempty"`
	FontIcon *BloodHoundGraphFontIcon   `json:"fontIcon,omitempty"`
	Image    string                     `json:"image,omitempty"`
	Label    *BloodHoundGraphLabel      `json:"label,omitempty"`
	Position string                     `json:"position,omitempty"` //Needs to be a compass direction (ne by default)
	Radius   int                        `json:"radius,omitempty"`
	Size     int                        `json:"size,omitempty"` //Defaults to 1
}

type BloodHoundGraphItemBorder struct {
	Color string `json:"color,omitempty"`
}

type BloodHoundGraphFontIcon struct {
	Color      string `json:"color,omitempty"`
	FontFamily string `json:"fontFamily,omitempty"`
	Text       string `json:"text,omitempty"`
}

type BloodHoundGraphLabel struct {
	Bold       bool   `json:"bold,omitempty"`
	Color      string `json:"color,omitempty"`
	FontFamily string `json:"fontFamily,omitempty"`
	Text       string `json:"text,omitempty"`
}

type BloodHoundGraphItem struct {
	Color  string                  `json:"color,omitempty"`
	Data   map[string]any          `json:"data,omitempty"`
	Fade   bool                    `json:"fade,omitempty"`
	Glyphs *[]BloodHoundGraphGlyph `json:"glyphs,omitempty"`
}

type BloodHoundGraphNodeBorder struct {
	Color     string `json:"color,omitempty"`
	LineStyle string `json:"lineStyle,omitempty"` //solid or dashed
	Width     int    `json:"width,omitempty"`
}

type BloodHoundGraphNodeCoords struct {
	Latitude  int `json:"lat,omitempty"`
	Longitude int `json:"lng,omitempty"`
}

type BloodHoundGraphNodeHalo struct {
	Color  string `json:"color,omitempty"`
	Radius int    `json:"radius,omitempty"`
	Width  int    `json:"width,omitempty"`
}

type BloodHoundGraphNodeLabel struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Bold            bool   `json:"bold,omitempty"`
	Center          bool   `json:"center,omitempty"`
	Color           string `json:"color,omitempty"`
	FontFamily      string `json:"fontFamily,omitempty"`
	FontSize        int    `json:"fontSize,omitempty"`
	Text            string `json:"text,omitempty"`
}

type BloodHoundGraphNode struct {
	*BloodHoundGraphItem
	Border      *BloodHoundGraphNodeBorder `json:"border,omitempty"`
	Coordinates *BloodHoundGraphNodeCoords `json:"coordinates,omitempty"`
	Cutout      bool                       `json:"cutout,omitempty"`
	FontIcon    *BloodHoundGraphFontIcon   `json:"fontIcon,omitempty"`
	Halos       *[]BloodHoundGraphNodeHalo `json:"halos,omitempty"`
	Image       string                     `json:"image,omitempty"`
	Label       *BloodHoundGraphNodeLabel  `json:"label,omitempty"`
	Shape       string                     `json:"shape,omitempty"`
	Size        int                        `json:"size,omitempty"`
}

type BloodHoundGraphLinkLabel struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Bold            bool   `json:"bold,omitempty"`
	Color           string `json:"color,omitempty"`
	FontFamily      string `json:"fontFamily,omitempty"`
	FontSize        int    `json:"fontSize,omitempty"`
	Text            string `json:"text,omitempty"`
}

type BloodHoundGraphLinkEnd struct {
	Arrow   bool                      `json:"arrow,omitempty"`
	BackOff int                       `json:"backOff,omitempty"`
	Color   string                    `json:"color,omitempty"`
	Glyphs  *[]BloodHoundGraphGlyph   `json:"glyphs,omitempty"`
	Label   *BloodHoundGraphLinkLabel `json:"label,omitempty"`
}

type BloodHoundGraphLinkFlow struct {
	Velocity int `json:"velocity,omitempty"`
}

type BloodHoundGraphLink struct {
	*BloodHoundGraphItem
	End1      *BloodHoundGraphLinkEnd   `json:"end1,omitempty"`
	End2      *BloodHoundGraphLinkEnd   `json:"end2,omitempty"`
	Flow      *BloodHoundGraphLinkFlow  `json:"flow,omitempty"`
	ID1       string                    `json:"id1,omitempty"`
	ID2       string                    `json:"id2,omitempty"`
	Label     *BloodHoundGraphLinkLabel `json:"label,omitempty"`
	LineStyle string                    `json:"lineStyle,omitempty"`
	Width     int                       `json:"width,omitempty"`
}

func (s *BloodHoundGraphNode) SetFontIcon(nodeKind string, schemaNodeKinds model.GraphSchemaNodeKindMap, schemalessNodeKinds model.CustomNodeKindMap) {
	if nodeKind == metaNodeKindLabel {
		s.Color = metaNodeColor

		if tier, ok := s.Data["admintier"]; ok {
			if tier.(int64) == 0 {
				s.Image = metaImageTierZero
			} else {
				s.Image = metaImageDefault
			}
		} else {
			s.Image = metaImageDefault
		}
	} else if nodeKindConfig, ok := schemaNodeKinds[nodeKind]; ok {
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: fmt.Sprintf("%s%s", fontAwesomePrefix, nodeKindConfig.Icon),
		}
		s.Color = nodeKindConfig.IconColor
	} else if schemalessNodeKindConfig, ok := schemalessNodeKinds[nodeKind]; ok {
		s.Data["nodetype"] = nodeKind
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: fmt.Sprintf("%s%s", fontAwesomePrefix, schemalessNodeKindConfig.Icon.Name),
		}
		s.Color = schemalessNodeKindConfig.Icon.Color
	} else {
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: defaultUnknownIcon,
		}
		s.Color = defaultIconColor
	}
}
