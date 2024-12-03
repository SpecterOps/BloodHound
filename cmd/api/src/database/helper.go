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

package database

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
	"strings"
)

func CheckError(tx *gorm.DB) error {
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}

	if tx.Error != nil {
		if strings.Contains(tx.Error.Error(), "duplicate key value violates unique constraint \"users_principal_name_key\"") {
			return fmt.Errorf("%w: %v", ErrDuplicateUserPrincipal, tx.Error)
		}
	}

	return tx.Error
}

// NullUUID returns a uuid.NullUUID struct i.e a UUID that can be null in pg
func NullUUID(value uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{
		UUID:  value,
		Valid: true,
	}
}
