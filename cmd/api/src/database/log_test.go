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

package database_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/database"
)

func TestGormLogAdapter_Info(t *testing.T) {
	t.Parallel()
	var (
		gormLogAdapter = database.GormLogAdapter{
			SlowQueryWarnThreshold:  time.Minute,
			SlowQueryErrorThreshold: time.Minute,
		}
	)

	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{})))

	expected := fmt.Sprintf(`msg="message %d %s %f"`, 1, "arg", 2.0)
	gormLogAdapter.Info(context.TODO(), "message %d %s %f", 1, "arg", 2.0)
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("gormLogAdapter output does not contain expected\nOutput:%sExpected:%s", buf.String(), expected)
	}
}
