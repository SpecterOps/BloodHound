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
export interface UserPreferences extends Record<string, unknown> {
    version: 1;
    darkMode: boolean;
    rememberDomain: boolean;
    preferredDomainId: string | null;
}

export const defaultUserPreferences: UserPreferences = {
    version: 1,
    darkMode: false,
    rememberDomain: false,
    preferredDomainId: null,
};

// Only the frontend interprets this versioned UTF-8 document. Unknown fields
// survive writes so another UI build's settings are not inadvertently removed.
export const decodeUserPreferences = (storage: ArrayBuffer): UserPreferences => {
    if (storage.byteLength === 0) return { ...defaultUserPreferences };
    const value: unknown = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(storage));
    if (
        !value ||
        typeof value !== 'object' ||
        Array.isArray(value) ||
        !('version' in value) ||
        value.version !== 1 ||
        !('darkMode' in value) ||
        typeof value.darkMode !== 'boolean' ||
        !('rememberDomain' in value) ||
        typeof value.rememberDomain !== 'boolean' ||
        !('preferredDomainId' in value) ||
        (value.preferredDomainId !== null && typeof value.preferredDomainId !== 'string')
    ) {
        throw new Error('These preferences cannot be read by this version of BloodHound.');
    }
    return value as UserPreferences;
};

export const encodeUserPreferences = (preferences: UserPreferences): ArrayBuffer => {
    const bytes = new TextEncoder().encode(JSON.stringify(preferences));
    if (bytes.byteLength > 65536) throw new Error('User preferences exceed the storage limit.');
    return bytes.buffer;
};
