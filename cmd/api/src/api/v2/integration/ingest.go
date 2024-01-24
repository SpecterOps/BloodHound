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

package integration

import (
	"encoding/json"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/fixtures"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func ingestPayload(t test.Controller, loader fixtures.Loader, fixturePath string) map[string]any {
	payload := map[string]any{}

	if err := json.Unmarshal(loader.Get(fixturePath), &payload); err != nil {
		t.Fatalf("Failed decoding ingest fixture %s: %v", fixturePath, err)
	}

	return payload
}

func (s *Context) ToggleFeatureFlag(name string) {
	require.Nil(s.TestCtrl, s.AdminClient().ToggleFeatureFlag(name))
}

func (s *Context) SendFileIngest(fixtures []string) {
	apiClient := s.AdminClient()

	if uploadJob, err := apiClient.CreateFileUploadTask(); err != nil {
		s.TestCtrl.Fatalf("Failed creating upload job: %v", err)
	} else {
		for _, fixtureName := range fixtures {
			var ingestData = ingestPayload(s.TestCtrl, s.FixtureLoader, fixtureName)

			if err := apiClient.SendFileUploadData(ingestData, uploadJob.ID); err != nil {
				s.TestCtrl.Fatalf("Failed sending ingest for fixture %s: %v", fixtureName, err)
			}
		}

		s.WaitForDatapipeIdle(90 * time.Second)

		originalStatusWrapper := s.GetDatapipeStatusWrapper()

		if err := apiClient.CompleteFileUpload(uploadJob.ID); err != nil {
			s.TestCtrl.Fatalf("Failed completing job %d: %v", uploadJob.ID, err)
		}

		s.WaitForDatapipeAnalysis(90*time.Second, originalStatusWrapper)
	}
}

func (s *Context) SendCompressedFileIngest(fixtures []string) {
	apiClient := s.AdminClient()

	if uploadJob, err := apiClient.CreateFileUploadTask(); err != nil {
		s.TestCtrl.Fatalf("Failed creating upload job: %v", err)
	} else {

		for _, fixtureName := range fixtures {
			jsonInput := s.FixtureLoader.GetReader(fixtureName)
			defer jsonInput.Close()

			if err := apiClient.SendCompressedFileUploadData(jsonInput, uploadJob.ID); err != nil {
				s.TestCtrl.Fatalf("Failed sending compressed ingest for fixture %s: %v", fixtureName, err)
			}
		}

		s.WaitForDatapipeIdle(90 * time.Second)

		originalStatusWrapper := s.GetDatapipeStatusWrapper()

		if err := apiClient.CompleteFileUpload(uploadJob.ID); err != nil {
			s.TestCtrl.Fatalf("Failed completing job %d: %v", uploadJob.ID, err)
		}

		s.WaitForDatapipeAnalysis(90*time.Second, originalStatusWrapper)
	}
}

func (s *Context) WaitForDatapipeIdle(timeout time.Duration) {
	start := time.Now()
	for {
		if status, err := s.AdminClient().GetDatapipeStatus(); err != nil {
			s.TestCtrl.Fatalf("Error getting datapipe status: %v", err)
		} else if status.Status == model.DatapipeStatusIdle {
			break
		} else if elapsed := time.Since(start); elapsed >= timeout {
			s.TestCtrl.Fatalf("Waited too long for datapipe to be idle. Waited %d seconds - current datapipe status is: %s", elapsed.Seconds(), status.Status)
		}
		time.Sleep(time.Second)
	}
}

func (s *Context) GetDatapipeStatusWrapper() model.DatapipeStatusWrapper {
	status, err := s.AdminClient().GetDatapipeStatus()

	if err != nil {
		s.TestCtrl.Fatalf("Failed getting datapipe status: %v", err)
	}

	return status
}

func (s *Context) WaitForDatapipeAnalysis(timeout time.Duration, originalWrapper model.DatapipeStatusWrapper) {
	start := time.Now()

	time.Sleep(time.Second * 2)
	for {
		if status, err := s.AdminClient().GetDatapipeStatus(); err != nil {
			s.TestCtrl.Fatalf("Error getting datapipe status: %v", err)
		} else if status.Status == model.DatapipeStatusIdle && status.LastCompleteAnalysisAt.After(originalWrapper.LastCompleteAnalysisAt) {
			break
		} else if elapsed := time.Since(start); elapsed >= timeout {
			s.TestCtrl.Fatalf("Waited too long for datapipe to finish. Waited %d seconds - current datapipe status is: %s", elapsed.Seconds(), status.Status)
		}

		time.Sleep(time.Second)
	}
}

type IngestAssertion func(testCtrl test.Controller, tx graph.Transaction)

func (s *Context) AssertIngest(assertion IngestAssertion) {
	graphDB := integration.OpenGraphDB(s.TestCtrl)
	defer graphDB.Close(s.ctx)

	require.Nil(s.TestCtrl, graphDB.ReadTransaction(s.ctx, func(tx graph.Transaction) error {
		assertion(s.TestCtrl, tx)
		return nil
	}), "Unexpected database error during reconciliation assertion")
}
