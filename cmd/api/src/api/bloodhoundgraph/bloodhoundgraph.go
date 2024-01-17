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

type BloodHoundGraphNodeDonutBorder struct {
	Color string `json:"color,omitempty"`
	Width int    `json:"width,omitempty"`
}

type BloodHoundGraphNodeDonutSegment struct {
	Color string `json:"color,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type BloodHoundGraphNodeDonut struct {
	Border   *BloodHoundGraphNodeDonutBorder    `json:"border,omitempty"`
	Segments *[]BloodHoundGraphNodeDonutSegment `json:"segments,omitempty"`
	Width    int                                `json:"number,omitempty"`
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

func (s *BloodHoundGraphNode) SetIcon(nType string) {
	switch nType {
	case "AZApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-window-restore",
		}
	case "AZVMScaleSet":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-server",
		}
	case "AZDevice":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-desktop",
		}
	case "AZFunctionApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-bolt",
		}
	case "AZGroup":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-users",
		}
	case "AZKeyVault":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-lock",
		}
	case "AZManagementGroup":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-sitemap",
		}
	case "AZResourceGroup":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-cube",
		}
	case "AZRole":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-clipboard-list",
		}
	case "AZServicePrincipal":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-robot",
		}
	case "AZSubscription":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-key",
		}
	case "AZTenant":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-cloud",
		}
	case "AZUser":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-user",
		}
	case "AZVM":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-desktop",
		}
	case "AZManagedCluster":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-cubes",
		}
	case "AZContainerRegistry":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-box-open",
		}
	case "AZWebApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-object-group",
		}
	case "AZLogicApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-sitemap",
		}
	case "AZAutomationAccount":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-cog",
		}
	case "User":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-user",
		}
	case "Group":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-users",
		}
	case "Computer":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-desktop",
		}
	case "Container":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-box",
		}
	case "Domain":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-globe",
		}
	case "OU":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-sitemap",
		}
	case "GPO":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-list",
		}
	case "AIACA":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-box",
		}
	case "RootCA":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-landmark",
		}
	case "EnterpriseCA":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-building",
		}
	case "NTAuthStore":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-store",
		}
	case "CertTemplate":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-id-card",
		}
	case "Meta":
		if tier, ok := s.Data["admintier"]; ok {
			if tier.(int64) == 0 {
				s.Image = "/ui/metat0.png"
			} else {
				s.Image = "/ui/meta.png"
			}
		} else {
			s.Image = "/ui/meta.png"
		}
	default:
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fa-question",
		}
	}
}

func (s *BloodHoundGraphNode) SetBackground(nType string) {
	switch nType {
	case "AZApp":
		s.BloodHoundGraphItem.Color = "#03FC84"
	case "AZVMScaleSet":
		s.BloodHoundGraphItem.Color = "#007CD0"
	case "AZDevice":
		s.BloodHoundGraphItem.Color = "#B18FCF"
	case "AZFunctionApp":
		s.BloodHoundGraphItem.Color = "#F4BA44"
	case "AZGroup":
		s.BloodHoundGraphItem.Color = "#F57C9B"
	case "AZKeyVault":
		s.BloodHoundGraphItem.Color = "#ED658C"
	case "AZManagementGroup":
		s.BloodHoundGraphItem.Color = "#BD93D8"
	case "AZResourceGroup":
		s.BloodHoundGraphItem.Color = "#89BD9E"
	case "AZRole":
		s.BloodHoundGraphItem.Color = "#ED8537"
	case "AZServicePrincipal":
		s.BloodHoundGraphItem.Color = "#C1D6D6"
	case "AZSubscription":
		s.BloodHoundGraphItem.Color = "#D2CCA1"
	case "AZTenant":
		s.BloodHoundGraphItem.Color = "#54F2F2"
	case "AZUser":
		s.BloodHoundGraphItem.Color = "#34D2EB"
	case "AZVM":
		s.BloodHoundGraphItem.Color = "#F9ADA0"
	case "AZManagedCluster":
		s.BloodHoundGraphItem.Color = "#326CE5"
	case "AZContainerRegistry":
		s.BloodHoundGraphItem.Color = "#0885D7"
	case "AZWebApp":
		s.BloodHoundGraphItem.Color = "#4696E9"
	case "AZLogicApp":
		s.BloodHoundGraphItem.Color = "#9EE047"
	case "AZAutomationAccount":
		s.BloodHoundGraphItem.Color = "#F4BA44"
	case "User":
		s.BloodHoundGraphItem.Color = "#17E625"
	case "Group":
		s.BloodHoundGraphItem.Color = "#DBE617"
	case "Computer":
		s.BloodHoundGraphItem.Color = "#E67873"
	case "Container":
		s.BloodHoundGraphItem.Color = "#F79A78"
	case "Domain":
		s.BloodHoundGraphItem.Color = "#17E6B9"
	case "OU":
		s.BloodHoundGraphItem.Color = "#FFAA00"
	case "GPO":
		s.BloodHoundGraphItem.Color = "#998EFD"
	case "AIACA":
		s.BloodHoundGraphItem.Color = "#9769F0"
	case "RootCA":
		s.BloodHoundGraphItem.Color = "#6968E8"
	case "EnterpriseCA":
		s.BloodHoundGraphItem.Color = "#4696E9"
	case "NTAuthStore":
		s.BloodHoundGraphItem.Color = "#D575F5"
	case "CertTemplate":
		s.BloodHoundGraphItem.Color = "#B153F3"
	case "Meta":
		s.BloodHoundGraphItem.Color = "#000"
	default:
		s.BloodHoundGraphItem.Color = "#EEE"
	}
}
