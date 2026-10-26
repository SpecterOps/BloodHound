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
import { cn } from '../utils';

export const textareaStyles = cn(
    'w-full resize-none rounded-md border border-[color:var(--textarea-border-default)] bg-[color:var(--textarea-fill)] px-3 py-2 text-main placeholder:text-[color:var(--input-placeholder-text)]',
    'enabled:hover:border-[color:var(--textarea-border-hover)]',
    'focus:outline-none focus:focus-ring focus:border-[color:var(--textarea-focus-border)] enabled:hover:focus:border-[color:var(--textarea-hover-focus-border)]',
    'disabled:cursor-not-allowed disabled:border-[color:var(--input-border-disabled)] disabled:bg-[color:var(--textarea-disabled-fill)] disabled:text-disabled disabled:placeholder:text-disabled disabled:opacity-100',
    'disabled:aria-[invalid=true]:border-[color:var(--input-disabled-invalid-border)]',
    'transition-[color,box-shadow] md:text-sm',
    'aria-[invalid=true]:border-status-error-main aria-[invalid=true]:hover:border-status-error-main aria-[invalid=true]:focus:border-status-error-main aria-[invalid=true]:enabled:hover:focus:border-status-error-main'
);
