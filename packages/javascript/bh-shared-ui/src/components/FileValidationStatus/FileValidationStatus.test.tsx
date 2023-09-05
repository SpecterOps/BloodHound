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

import { render } from '../../test-utils';
import { FileForIngest, FileStatus } from '../FileUploadDialog/types';
import FileValidationStatus from './FileValidationStatus';

describe('FileValidationStatus', () => {
    const fileForIngest: FileForIngest = {
        file: new File([''], 'mock.json', { type: 'application/json' }),
        status: FileStatus.READY,
    };

    it('should display a check for files that are ready to upload', () => {
        const container = render(<FileValidationStatus file={fileForIngest} />);
        expect(container.getByText('check')).toBeInTheDocument();
    });

    it('should display a loading spinner for files that are uploading', () => {
        fileForIngest.status = FileStatus.UPLOADING;
        const container = render(<FileValidationStatus file={fileForIngest} />);
        expect(container.getByText('arrows-rotate')).toBeInTheDocument();
    });

    it('should display an x and error messages for files that have errors', () => {
        fileForIngest.status = FileStatus.READY;
        fileForIngest.errors = ['test error'];
        const container = render(<FileValidationStatus file={fileForIngest} />);
        expect(container.getByText('xmark')).toBeInTheDocument();
        expect(container.getByText('test error')).toBeInTheDocument();
    });
});
