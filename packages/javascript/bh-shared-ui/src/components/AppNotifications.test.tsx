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

import { SNACKBAR_DURATION } from '../constants';
import { useNotifications } from '../providers';
import { act, renderHook } from '../test-utils';

const message = 'This is a notification';
const messageKey = 'messageKey';

describe('AppNotifications', () => {
    it('adds a notification - checks values', () => {
        const hook = renderHook(() => useNotifications());

        act(() => hook.result.current.addNotification(message, messageKey, { autoHideDuration: SNACKBAR_DURATION }));

        const result = hook.result.current.notifications[0];
        expect(result.message).toBe(message);
        expect(result.key).toBe(messageKey);
        expect(result.options.autoHideDuration).toBe(SNACKBAR_DURATION);
    });

    it('dismisses a notification', () => {
        const hook = renderHook(() => useNotifications());

        act(() => hook.result.current.addNotification(message, messageKey));
        expect(hook.result.current.notifications).toHaveLength(1);
        expect(hook.result.current.notifications[0].dismissed).toBeFalsy;

        act(() => hook.result.current.dismissNotification(messageKey));
        expect(hook.result.current.notifications[0].dismissed).toBeTruthy;
    });

    it('removes a notification', () => {
        const hook = renderHook(() => useNotifications());

        act(() => hook.result.current.addNotification(message, messageKey));
        expect(hook.result.current.notifications).toHaveLength(1);

        act(() => hook.result.current.removeNotification(messageKey));
        expect(hook.result.current.notifications).toHaveLength(0);
    });
});
