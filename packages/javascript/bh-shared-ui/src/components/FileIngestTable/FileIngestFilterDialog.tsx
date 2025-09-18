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
import { useMemo, useState } from 'react';
import { useObjectState } from '../../hooks';
import { typedEntries } from '../../utils';
import { AppIcon } from '../AppIcon';
import { DateRangeChange, DateRangeInputs } from '../DateRangeInputs';
import { StatusSelect } from '../StatusSelect';
import { UserSelect } from './UserSelect';

type Props = {
    onConfirm: (filters: any) => void;
};

export const FileIngestFilterDialog: React.FC<Props> = ({ onConfirm }) => {
    const [areDatesValid, setAreDatesValid] = useState(true);

    // Manages filter state while selecting option in dialog
    // Sent to table via onConfirm
    const filters = useObjectState<any>({});

    const clearFilters = () => filters.setState({});

    const selectStatus = (status: string) => {
        if (status === '-none-') {
            filters.deleteKeys('status');
        } else {
            filters.applyState({ status: parseInt(status, 10) });
        }
    };

    const setDateRange = (changed: DateRangeChange) => {
        const entries = typedEntries(changed);
        if (entries.length === 0) return;
        const [key, value] = entries[0];

        if (value === undefined) {
            filters.deleteKeys(key);
        } else {
            filters.applyState(changed);
        }
    };

    const selectUser = (client_id: string) => {
        if (client_id === '-none-') {
            filters.deleteKeys('client_id');
        } else {
            filters.applyState({ client_id });
        }
    };

    const isConfirmDisabled = useMemo(() => {
        return Object.keys(filters.state).length === 0 || !areDatesValid;
    }, [areDatesValid, filters.state]);

    return (
        <Dialog>
            <DialogTrigger asChild>
                <Button data-testid='finished_jobs_log-open_filter_dialog' variant='icon'>
                    <AppIcon.FilterOutline size={22} />
                </Button>
            </DialogTrigger>

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
                        <StatusSelect
                            status={filters.state.status}
                            statusOptions={['Complete', 'Running', 'Failed']}
                            onSelect={selectStatus}
                        />
                    </DialogDescription>

                    <DialogDescription asChild>
                        <DateRangeInputs
                            end={filters.state.end_time}
                            onChange={setDateRange}
                            onValidation={setAreDatesValid}
                            start={filters.state.start_time}
                        />
                    </DialogDescription>

                    <DialogDescription asChild>
                        <UserSelect client={filters.state.client_id} onSelect={selectUser} />
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
                                onClick={() => onConfirm(filters.state)}
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
