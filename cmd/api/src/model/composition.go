// Copyright 2025 Specter Ops, Inc.
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

package model

// These structs were created for the new edge composition model, but are being saved for later use, since it doesn't work with our current post implementation
type EdgeCompositionEdge struct {
	PostProcessedEdgeID int64
	CompositionEdgeID   int64

	BigSerial
}

type EdgeCompositionEdges []EdgeCompositionEdge

type EdgeCompositionNode struct {
	PostProcessedEdgeID int64
	CompositionNodeID   int64

	BigSerial
}

type EdgeCompositionNodes []EdgeCompositionNode
