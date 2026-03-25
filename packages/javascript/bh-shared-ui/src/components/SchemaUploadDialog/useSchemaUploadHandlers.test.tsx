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

import { act, renderHook } from '../../test-utils';
import { extensionsKeys } from '../../hooks';
import { useSchemaUploadHandlers } from './useSchemaUploadHandlers';

const invalidateQueriesMock = vi.hoisted(() => vi.fn());
const mutateMock = vi.hoisted(() => vi.fn());

vi.mock('react-query', async () => {
    const actual = await vi.importActual<typeof import('react-query')>('react-query');

    return {
        ...actual,
        useMutation: () => ({ mutate: mutateMock }),
        useQueryClient: () => ({ invalidateQueries: invalidateQueriesMock }),
    };
});

const createFileList = (file: File): FileList =>
    ({
        0: file,
        length: 1,
        item: (index: number) => (index === 0 ? file : null),
    }) as unknown as FileList;

describe('useSchemaUploadHandlers', () => {
    afterEach(() => {
        invalidateQueriesMock.mockReset();
        mutateMock.mockReset();
    });

    it('invalidates the extensions query key when upload succeeds', () => {
        mutateMock.mockImplementation((_variables, options: { onSuccess?: () => void } = {}) => {
            options.onSuccess?.();
        });

        const testFile = new File([JSON.stringify({ value: 'test' })], 'test.json', { type: 'application/json' });
        const { result } = renderHook(() => useSchemaUploadHandlers());

        act(() => result.current.handleFileDrop(createFileList(testFile)));
        act(() => result.current.handleUpload());

        expect(invalidateQueriesMock).toHaveBeenCalledWith({ queryKey: extensionsKeys.all });
    });
});