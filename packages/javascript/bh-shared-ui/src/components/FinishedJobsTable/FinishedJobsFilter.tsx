import {
    Button,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    DialogTrigger,
    Form,
} from '@bloodhoundenterprise/doodleui';
import { noop } from 'lodash';
import React from 'react';
import { useForm } from 'react-hook-form';
import { AppIcon } from '../AppIcon';
import { DataCollectedSelect } from './DataCollectedSelect';
import { StatusSelect } from './StatusSelect';

export const FinishedJobsFilter: React.FC = () => {
    const form = useForm({});

    const applyFilter = () => {
        console.log(form.getValues());
    };

    return (
        <Dialog>
            <DialogTrigger asChild>
                <div className='mb-4 text-right'>
                    <Button
                        variant='icon'
                        // data-testid='posture_attack_paths-filter_dialog_trigger'
                        // disabled={disableTrigger}
                    >
                        <AppIcon.FilterOutline size={22} />
                    </Button>
                </div>
            </DialogTrigger>

            <DialogPortal>
                <DialogContent>
                    <DialogTitle className='flex justify-between items-center'>
                        Filter
                        <Button variant='text' className='font-normal py-0 h-fit' onClick={noop}>
                            Clear All
                        </Button>
                    </DialogTitle>

                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(applyFilter)}>
                            <DialogDescription asChild>
                                <div className='flex gap-10'>
                                    <StatusSelect control={form.control} />
                                    <DataCollectedSelect control={form.control} />
                                </div>
                            </DialogDescription>

                            <DialogActions>
                                <DialogClose asChild>
                                    <Button
                                        variant='text'
                                        className='pr-0'
                                        // data-testid='posture_attack_paths-filter_dialog_close'
                                    >
                                        Cancel
                                    </Button>
                                </DialogClose>
                                <DialogClose asChild>
                                    <Button variant='text' className='text-primary' type='submit'>
                                        Confirm
                                    </Button>
                                </DialogClose>
                            </DialogActions>
                        </form>
                    </Form>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};
