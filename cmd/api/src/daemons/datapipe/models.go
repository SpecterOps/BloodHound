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

package datapipe

import (
	"encoding/json"

	"github.com/bloodhoundad/azurehound/v2/enums"
	"github.com/specterops/bloodhound/ein"
)

type ConvertedData struct {
	NodeProps []ein.IngestibleNode
	RelProps  []ein.IngestibleRelationship
}

func (s *ConvertedData) Clear() {
	s.NodeProps = s.NodeProps[:0]
	s.RelProps = s.RelProps[:0]
}

type ConvertedGroupData struct {
	NodeProps              []ein.IngestibleNode
	RelProps               []ein.IngestibleRelationship
	DistinguishedNameProps []ein.IngestibleRelationship
}

func (s *ConvertedGroupData) Clear() {
	s.NodeProps = s.NodeProps[:0]
	s.RelProps = s.RelProps[:0]
	s.DistinguishedNameProps = s.DistinguishedNameProps[:0]
}

type ConvertedSessionData struct {
	SessionProps []ein.IngestibleSession
}

func (s *ConvertedSessionData) Clear() {
	s.SessionProps = s.SessionProps[:0]
}

type AzureBase struct {
	Kind enums.Kind      `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type ConvertedAzureData struct {
	NodeProps   []ein.IngestibleNode
	RelProps    []ein.IngestibleRelationship
	OnPremNodes []ein.IngestibleNode
}

func (s *ConvertedAzureData) Clear() {
	s.NodeProps = s.NodeProps[:0]
	s.RelProps = s.RelProps[:0]
	s.OnPremNodes = s.OnPremNodes[:0]
}
