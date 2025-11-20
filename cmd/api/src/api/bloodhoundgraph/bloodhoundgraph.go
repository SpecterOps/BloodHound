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

//TODO: Move styling responsibilities to the UI or move shared styling definitions to a cue file to generate from one source of truth

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

func (s *BloodHoundGraphNode) SetNodeStyle(nType string) {
	s.SetIcon(nType)
	s.SetBackground(nType)
}

func (s *BloodHoundGraphNode) SetIcon(nType string) {
	switch nType {
	case "AZApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-window-restore",
		}
	case "AZVMScaleSet":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-server",
		}
	case "AZDevice":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-desktop",
		}
	case "AZFunctionApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-bolt",
		}
	case "AZGroup":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-users",
		}
	case "AZKeyVault":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-lock",
		}
	case "AZManagementGroup":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-sitemap",
		}
	case "AZResourceGroup":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-cube",
		}
	case "AZRole":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-clipboard-list",
		}
	case "AZServicePrincipal":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-robot",
		}
	case "AZSubscription":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-key",
		}
	case "AZTenant":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-cloud",
		}
	case "AZUser":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-user",
		}
	case "AZVM":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-desktop",
		}
	case "AZManagedCluster":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-cubes",
		}
	case "AZContainerRegistry":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-box-open",
		}
	case "AZWebApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-object-group",
		}
	case "AZLogicApp":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-sitemap",
		}
	case "AZAutomationAccount":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-cog",
		}
	case "User":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-user",
		}
	case "Group":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-users",
		}
	case "Computer":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-desktop",
		}
	case "Container":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-box",
		}
	case "Domain":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-globe",
		}
	case "OU":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-sitemap",
		}
	case "GPO":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-list",
		}
	case "AIACA":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-arrows-left-right-to-line",
		}
	case "RootCA":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-landmark",
		}
	case "EnterpriseCA":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-building",
		}
	case "NTAuthStore":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-store",
		}
	case "CertTemplate":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-id-card",
		}
	case "IssuancePolicy":
		s.FontIcon = &BloodHoundGraphFontIcon{
			Text: "fas fa-clipboard-check",
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
			Text: "fas fa-question",
		}
	}
}

func (s *BloodHoundGraphNode) SetBackground(nType string) {
	switch nType {
	case "AZApp":
		s.Color = "#03FC84"
	case "AZVMScaleSet":
		s.Color = "#007CD0"
	case "AZDevice":
		s.Color = "#B18FCF"
	case "AZFunctionApp":
		s.Color = "#F4BA44"
	case "AZGroup":
		s.Color = "#F57C9B"
	case "AZKeyVault":
		s.Color = "#ED658C"
	case "AZManagementGroup":
		s.Color = "#BD93D8"
	case "AZResourceGroup":
		s.Color = "#89BD9E"
	case "AZRole":
		s.Color = "#ED8537"
	case "AZServicePrincipal":
		s.Color = "#C1D6D6"
	case "AZSubscription":
		s.Color = "#D2CCA1"
	case "AZTenant":
		s.Color = "#54F2F2"
	case "AZUser":
		s.Color = "#34D2EB"
	case "AZVM":
		s.Color = "#F9ADA0"
	case "AZManagedCluster":
		s.Color = "#326CE5"
	case "AZContainerRegistry":
		s.Color = "#0885D7"
	case "AZWebApp":
		s.Color = "#4696E9"
	case "AZLogicApp":
		s.Color = "#9EE047"
	case "AZAutomationAccount":
		s.Color = "#F4BA44"
	case "User":
		s.Color = "#17E625"
	case "Group":
		s.Color = "#DBE617"
	case "Computer":
		s.Color = "#E67873"
	case "Container":
		s.Color = "#F79A78"
	case "Domain":
		s.Color = "#17E6B9"
	case "OU":
		s.Color = "#FFAA00"
	case "GPO":
		s.Color = "#998EFD"
	case "AIACA":
		s.Color = "#9769F0"
	case "RootCA":
		s.Color = "#6968E8"
	case "EnterpriseCA":
		s.Color = "#4696E9"
	case "NTAuthStore":
		s.Color = "#D575F5"
	case "CertTemplate":
		s.Color = "#B153F3"
	case "Meta":
		s.Color = "#000"
	default:
		s.Color = "#EEE"
	}
}
