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
import { noop } from 'lodash';
import React, { useState } from 'react';
import { useObjectState } from '../../hooks';
import { type FinishedJobsFilters } from '../../utils';
import { AppIcon } from '../AppIcon';

type Props = {
    onConfirm: (filters: FinishedJobsFilters) => void;
};

export const FinishedJobsFilter: React.FC<Props> = () => {
    const [isConfirmDisabled] = useState(true);
    const { setState: setFilters } = useObjectState<FinishedJobsFilters>({});

    const clearFilters = () => setFilters({});

    return (
        <Dialog>
            <div className='mb-4 text-right'>
                <DialogTrigger asChild>
                    <Button variant='icon' data-testid='finished_jobs_log-open_filter_dialog'>
                        <AppIcon.FilterOutline size={22} />
                    </Button>
                </DialogTrigger>
            </div>

            <DialogPortal>
                <DialogOverlay blurBackground />

                <DialogContent aria-describedby='Finished Jobs Log dialog'>
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
                            {/* TODO: BED-6404 */}
                            <span>Status Select</span>
                            <span>Collection Select</span>
                        </div>
                    </DialogDescription>

                    <DialogDescription asChild>
                        {/* TODO: BED-6405 */}
                        <span>Date Range Inputs</span>
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
                                // TODO: BED-6407
                                onClick={() => noop}
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
