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
import { mockPZPathParams } from '../../mocks/factories/privilegeZones';
import { act, renderHook } from '../../test-utils';
import { apiClient } from '../../utils/api';
import * as usePZParams from '../usePZParams/usePZPathParams';
import { useObjectCounts } from './useObjectCounts';

const usePZPathParamsSpy = vi.spyOn(usePZParams, 'usePZPathParams');
const getAssetGroupTagMembersCountSpy = vi.spyOn(apiClient, 'getAssetGroupTagMembersCount');
const getAssetGroupTagRuleMembersCountSpy = vi.spyOn(apiClient, 'getAssetGroupTagRuleMembersCount');

describe('useObjectCounts', () => {
    it('returns Tag Object counts when there is a Tag selected but no Rule selected (ruleId is undefined from path params)', async () => {
        usePZPathParamsSpy.mockReturnValue({ ...mockPZPathParams, tagId: '1', ruleId: undefined });

        await act(async () => {
            renderHook(() => useObjectCounts());
        });

        expect(getAssetGroupTagMembersCountSpy).toBeCalled();
        expect(getAssetGroupTagRuleMembersCountSpy).not.toBeCalled();
    });

    it('returns Rule Object counts when there is a Tag and a Rule selected (tagId and ruleId are in path params)', async () => {
        usePZPathParamsSpy.mockReturnValue({ ...mockPZPathParams, tagId: '1', ruleId: '2' });

        await act(async () => {
            renderHook(() => useObjectCounts());
        });
        expect(getAssetGroupTagMembersCountSpy).not.toBeCalled();
        expect(getAssetGroupTagRuleMembersCountSpy).toBeCalled();
    });

    it('makes no count request when neither a Tag or Rule is selected (tagId and ruleId are undefined from path params', async () => {
        //@ts-expect-error
        usePZPathParamsSpy.mockReturnValue({ ...mockPZPathParams, tagId: undefined, ruleId: '2' });

        await act(async () => {
            renderHook(() => useObjectCounts());
        });
        expect(getAssetGroupTagMembersCountSpy).not.toBeCalled();
        expect(getAssetGroupTagRuleMembersCountSpy).not.toBeCalled();
    });
});
