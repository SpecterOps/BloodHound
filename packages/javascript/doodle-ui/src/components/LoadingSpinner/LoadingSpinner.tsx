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

import { faCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon, type FontAwesomeIconProps } from '@fortawesome/react-fontawesome';
import { cn } from '../utils';
import { loadingSpinnerClasses, loadingSpinnerRootClasses } from './LoadingSpinner.styles';

function LoadingSpinner({ className, ...props }: Omit<FontAwesomeIconProps, 'ref' | 'icon'>) {
    return (
        <span role='progressbar' aria-label={props['aria-label'] ?? 'Loading'} className={loadingSpinnerRootClasses}>
            <FontAwesomeIcon icon={faCircle} className={cn(loadingSpinnerClasses, className)} {...props} />
        </span>
    );
}

export { LoadingSpinner };
