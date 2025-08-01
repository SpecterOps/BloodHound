import { useState } from 'react';

import {
    Button,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
} from '@bloodhoundenterprise/doodleui';
import { useNavigate } from 'react-router-dom';
import { useTagsQuery } from '../../../hooks';
type TagToZoneDialogProps = {
    dialogOpen: boolean;
    selectedQuery: any;
    setDialogOpen: (isOpen: boolean) => void;
};

const TagToZoneDialog = (props: TagToZoneDialogProps) => {
    const { dialogOpen, selectedQuery, setDialogOpen } = props;
    const navigate = useNavigate();

    const AssetGroupTagTypeTier = 1 as const;
    // export const AssetGroupTagTypeLabel = 2 as const;
    // export const AssetGroupTagTypeOwned = 3 as const;

    const tiersQuery = useTagsQuery((tag) => tag.type === AssetGroupTagTypeTier);
    const zones = tiersQuery.data;
    // console.log(zones);

    const [zone, setZone] = useState('');

    const handleValueChange = (val: any) => {
        console.log(`select value = ${val}`);
        setZone(val);
    };

    const onContinue = () => {
        navigate(`/zone-management/save/tier/${zone}/selector`, { state: selectedQuery });
    };

    return (
        <Dialog open={dialogOpen} onOpenChange={() => setDialogOpen(false)}>
            <DialogPortal>
                <DialogContent
                    DialogOverlayProps={{
                        blurBackground: false,
                    }}
                    maxWidth='sm'>
                    <DialogTitle>Tag Results to Zone</DialogTitle>

                    <DialogDescription>
                        Pick a zone to create a new selector. All assets returned by the query will be added to your
                        selector.
                    </DialogDescription>

                    <Select onValueChange={handleValueChange}>
                        <SelectTrigger className='w-60'>
                            <SelectValue placeholder='Select Zone' />
                        </SelectTrigger>
                        <SelectPortal>
                            <SelectContent>
                                {zones?.map((zone) => {
                                    return (
                                        <SelectItem key={zone.id} value={zone.id.toString()}>
                                            {zone.name}
                                        </SelectItem>
                                    );
                                })}
                            </SelectContent>
                        </SelectPortal>
                    </Select>

                    <DialogActions className='flex justify-end gap-4'>
                        <DialogClose asChild>
                            <Button variant='secondary'>Cancel</Button>
                        </DialogClose>

                        <Button disabled={!zone} onClick={onContinue}>
                            Continue
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default TagToZoneDialog;
