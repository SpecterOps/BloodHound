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

package neo4j

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/specterops/bloodhound/dawgs/graph"
	"time"
)

func AsTime(value any) (time.Time, error) {
	switch typedValue := value.(type) {
	case dbtype.Time:
		return typedValue.Time(), nil

	case dbtype.LocalTime:
		return typedValue.Time(), nil

	case dbtype.Date:
		return typedValue.Time(), nil

	case dbtype.LocalDateTime:
		return typedValue.Time(), nil

	default:
		return graph.AsTime(value)
	}
}

func mapValue(rawValue, target any) (bool, error) {
	switch typedTarget := target.(type) {
	case *time.Time:
		if value, err := AsTime(rawValue); err != nil {
			return false, err
		} else {
			*typedTarget = value
		}

	case *dbtype.Relationship:
		if value, typeOK := rawValue.(dbtype.Relationship); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Relationship", rawValue)
		} else {
			*typedTarget = value
		}

	case *graph.Relationship:
		if value, typeOK := rawValue.(dbtype.Relationship); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Relationship", rawValue)
		} else {
			*typedTarget = *newRelationship(value)
		}

	case *dbtype.Node:
		if value, typeOK := rawValue.(dbtype.Node); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Node", rawValue)
		} else {
			*typedTarget = value
		}

	case *graph.Node:
		if value, typeOK := rawValue.(dbtype.Node); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Node", rawValue)
		} else {
			*typedTarget = *newNode(value)
		}

	case *graph.Path:
		if value, typeOK := rawValue.(dbtype.Path); !typeOK {
			return false, fmt.Errorf("unexecpted type %T will not negotiate to *dbtype.Path", rawValue)
		} else {
			*typedTarget = newPath(value)
		}

	default:
		return false, nil
	}

	return true, nil
}

func NewValueMapper(values []any) graph.ValueMapper {
	return graph.NewValueMapper(values, mapValue)
}
