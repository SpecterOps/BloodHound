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

/** For use when value is expected but not yet available */
export type Maybe<T> = T | null | undefined;

/** Exclusive OR (XOR) type utility. */
export type XOR<T, U> =
    | (T & { [K in Exclude<keyof U, keyof T>]?: never })
    | (U & { [K in Exclude<keyof T, keyof U>]?: never });
