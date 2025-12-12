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
package opengraphschema

import (
	"context"
	"reflect"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func TestNewOpenGraphSchemaService(t *testing.T) {
	type args struct {
		openGraphSchemaRepository OpenGraphSchemaRepository
	}
	tests := []struct {
		name string
		args args
		want *OpenGraphSchemaService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOpenGraphSchemaService(tt.args.openGraphSchemaRepository); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOpenGraphSchemaService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenGraphSchemaService_UpsertGraphSchemaExtension(t *testing.T) {
	type fields struct {
		openGraphSchemaRepository OpenGraphSchemaRepository
	}
	type args struct {
		ctx         context.Context
		graphSchema model.GraphSchema
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OpenGraphSchemaService{
				openGraphSchemaRepository: tt.fields.openGraphSchemaRepository,
			}
			if err := o.UpsertGraphSchemaExtension(tt.args.ctx, tt.args.graphSchema); (err != nil) != tt.wantErr {
				t.Errorf("UpsertGraphSchemaExtension() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenGraphSchemaService_handleEdgeKindDiffActions(t *testing.T) {
	type fields struct {
		openGraphSchemaRepository OpenGraphSchemaRepository
	}
	type args struct {
		ctx         context.Context
		extensionId int32
		actions     MapDiffActions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OpenGraphSchemaService{
				openGraphSchemaRepository: tt.fields.openGraphSchemaRepository,
			}
			if err := o.handleEdgeKindDiffActions(tt.args.ctx, tt.args.extensionId, tt.args.actions); (err != nil) != tt.wantErr {
				t.Errorf("handleEdgeKindDiffActions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenGraphSchemaService_handleNodeKindDiffActions(t *testing.T) {
	type fields struct {
		openGraphSchemaRepository OpenGraphSchemaRepository
	}
	type args struct {
		ctx         context.Context
		extensionId int32
		actions     MapDiffActions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OpenGraphSchemaService{
				openGraphSchemaRepository: tt.fields.openGraphSchemaRepository,
			}
			if err := o.handleNodeKindDiffActions(tt.args.ctx, tt.args.extensionId, tt.args.actions); (err != nil) != tt.wantErr {
				t.Errorf("handleNodeKindDiffActions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenGraphSchemaService_handlePropertyDiffActions(t *testing.T) {
	type fields struct {
		openGraphSchemaRepository OpenGraphSchemaRepository
	}
	type args struct {
		ctx         context.Context
		extensionId int32
		actions     MapDiffActions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OpenGraphSchemaService{
				openGraphSchemaRepository: tt.fields.openGraphSchemaRepository,
			}
			if err := o.handlePropertyDiffActions(tt.args.ctx, tt.args.extensionId, tt.args.actions); (err != nil) != tt.wantErr {
				t.Errorf("handlePropertyDiffActions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_convertGraphSchemaEdgeKinds(t *testing.T) {
	type args struct {
		src *model.GraphSchemaEdgeKind
		dst *model.GraphSchemaEdgeKind
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertGraphSchemaEdgeKinds(tt.args.src, tt.args.dst)
		})
	}
}

func Test_convertGraphSchemaNodeKinds(t *testing.T) {
	type args struct {
		src *model.GraphSchemaNodeKind
		dst *model.GraphSchemaNodeKind
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertGraphSchemaNodeKinds(tt.args.src, tt.args.dst)
		})
	}
}

func Test_convertGraphSchemaProperties(t *testing.T) {
	type args struct {
		src *model.GraphSchemaProperty
		dst *model.GraphSchemaProperty
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertGraphSchemaProperties(tt.args.src, tt.args.dst)
		})
	}
}

func Test_validateGraphSchemModel(t *testing.T) {
	type args struct {
		graphSchema model.GraphSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateGraphSchemModel(tt.args.graphSchema); (err != nil) != tt.wantErr {
				t.Errorf("validateGraphSchemModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
