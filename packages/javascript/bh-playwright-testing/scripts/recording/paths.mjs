// Copyright 2026 Specter Ops, Inc.
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

import path from 'node:path';

export function recordingFiles(uiDirectory, name) {
    return {
        spec: path.join(uiDirectory, 'tests', 'e2e', `${name}.spec.ts`),
        har: path.join(uiDirectory, 'tests', 'e2e', `${name}.har`),
    };
}

export function temporaryRecordingPaths(uiDirectory, temporaryDirectory) {
    return {
        temporaryDirectory,
        storageState: path.join(uiDirectory, 'playwright', '.auth', 'user-light.json'),
        rawSource: path.join(temporaryDirectory, 'recording.ts'),
        rawHar: path.join(temporaryDirectory, 'recording.har'),
    };
}
