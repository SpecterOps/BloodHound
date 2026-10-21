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

export const textareaStyles = [
    'w-full resize-none rounded-md border border-textarea-border-default bg-textarea-fill px-3 py-2 text-main placeholder:text-input-placeholder-text',
    'enabled:hover:border-textarea-border-hover',
    'focus:outline-none focus:focus-ring dark:focus:border-textarea-border-default',
    'disabled:cursor-not-allowed disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-text-disabled disabled:placeholder:text-text-disabled disabled:opacity-100 dark:disabled:bg-input-fill-disabled-dark',
    'transition-[color,box-shadow] md:text-sm',
    'aria-[invalid=true]:border-status-error-main aria-[invalid=true]:hover:border-status-error-main aria-[invalid=true]:focus:border-status-error-main dark:aria-[invalid=true]:focus:border-status-error-main',
];
