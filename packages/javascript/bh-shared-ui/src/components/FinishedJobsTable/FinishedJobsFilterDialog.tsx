// Copyright 2025 Specter Ops, Inc.
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

import {
    Button,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogOverlay,
    DialogPortal,
    DialogTitle,
    DialogTrigger,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
import React, { useState } from 'react';
import { useObjectState } from '../../hooks';
import { getCollectionState, isCollectionKey, typedEntries, type FinishedJobsFilter } from '../../utils';
import { AppIcon } from '../AppIcon';
import { DataCollectedSelect } from './DataCollectedSelect';
import { DateRangeChange, DateRangeInputs } from './DateRangeInputs';
import { SELECT_NONE, StatusSelect } from './StatusSelect';

type Props = {
    onConfirm: (filters: FinishedJobsFilter) => void;
};

// TODO: BED-6407 - add onConfirm prop, executed when `Confirm` clicked
export const FinishedJobsFilterDialog: React.FC<Props> = () => {
    // TODO: BED-6407 - Disable confirm when range has validation error
    const [isConfirmDisabled] = useState(true);
    const filters = useObjectState<FinishedJobsFilter>({});

    const clearFilters = () => filters.setState({});

    const selectStatus = (status: string) => {
        if (status === SELECT_NONE) {
            filters.deleteKeys('status');
        } else {
            filters.applyState({ status: status });
        }
    };

    const toggleDataCollections = (key: string) => {
        if (!isCollectionKey(key)) return;

        if (key in filters.state) {
            filters.deleteKeys(key);
        } else {
            filters.applyState({ [key]: true });
        }
    };

    const setDateRange = (changed: DateRangeChange) => {
        const [key, value] = typedEntries(changed)[0];

        if (value === undefined) {
            filters.deleteKeys(key);
        } else {
            filters.applyState(changed);
        }
    };

    return (
        <Dialog>
            <div className='mb-4 text-right'>
                <DialogTrigger asChild>
                    <Button data-testid='finished_jobs_log-open_filter_dialog' variant='icon'>
                        <AppIcon.FilterOutline size={22} />
                    </Button>
                </DialogTrigger>
            </div>

            <DialogPortal>
                <DialogOverlay blurBackground />

                <DialogContent>
                    <DialogTitle className='flex justify-between items-center'>
                        Filter
                        <Button variant='text' className='font-normal p-0 h-fit' onClick={clearFilters}>
                            Clear All
                        </Button>
                    </DialogTitle>

                    <VisuallyHidden asChild>
                        <DialogDescription>Finished Jobs Log filters</DialogDescription>
                    </VisuallyHidden>

                    {/* Multiple Descriptions ensures that Dialog gaps still apply */}
                    <DialogDescription asChild>
                        <div className='flex gap-10'>
                            <StatusSelect status={filters.state.status} onSelect={selectStatus} />

                            <DataCollectedSelect
                                enabledCollections={getCollectionState(filters.state)}
                                onSelect={toggleDataCollections}
                            />
                        </div>
                    </DialogDescription>

                    <DialogDescription asChild>
                        <DateRangeInputs
                            start={filters.state.start_time}
                            end={filters.state.end_time}
                            onChange={setDateRange}
                        />
                    </DialogDescription>

                    <DialogDescription asChild>
                        {/* TODO: BED-6406 */}
                        <span>Client Select</span>
                    </DialogDescription>

                    <DialogActions>
                        <DialogClose asChild>
                            <Button
                                className='pr-0'
                                data-testid='finished_jobs_log-filter_dialog_close'
                                type='button'
                                variant='text'>
                                Cancel
                            </Button>
                        </DialogClose>
                        <DialogClose asChild>
                            <Button
                                className='text-primary'
                                data-testid='finished_jobs_log-filter_dialog_confirm'
                                disabled={isConfirmDisabled}
                                // TODO: BED-6407 - apply filters on click
                                // onClick={() => null}
                                type='submit'
                                variant='text'>
                                Confirm
                            </Button>
                        </DialogClose>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};
