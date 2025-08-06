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
import { QueryLineItem } from '../../../types';

type TagToZoneDialogProps = {
    dialogOpen: boolean;
    selectedQuery: QueryLineItem | undefined;
    isLabel: boolean;
    setDialogOpen: (isOpen: boolean) => void;
};

const TagToZoneDialog = (props: TagToZoneDialogProps) => {
    const { dialogOpen, selectedQuery, isLabel, setDialogOpen } = props;
    const navigate = useNavigate();

    const AssetGroupTagTypeTier = 1 as const;
    const AssetGroupTagTypeLabel = 2 as const;

    const tiersQuery = useTagsQuery((tag) => tag.type === AssetGroupTagTypeTier);
    const zones = tiersQuery.data;

    const labelsQuery = useTagsQuery((tag) => tag.type === AssetGroupTagTypeLabel);
    const labels = labelsQuery.data;

    const [zone, setZone] = useState('');
    const [label, setLabel] = useState('');

    const handleValueChange = (val: any) => {
        isLabel ? setLabel(val) : setZone(val);
    };

    const onContinue = () => {
        //TODO - use the const for this path
        if (isLabel) {
            navigate(`/zone-management/save/label/${label}/selector`, { state: selectedQuery });
        } else {
            navigate(`/zone-management/save/tier/${zone}/selector`, { state: selectedQuery });
        }
    };

    const title = isLabel ? 'label' : 'zone';

    const description = `Pick a ${title} to create a new selector. All assets returned by the query will be added to your selector.`;

    const continueDisabled = (isLabel && !label) || (!isLabel && !zone);

    return (
        <Dialog open={dialogOpen} onOpenChange={() => setDialogOpen(false)}>
            <DialogPortal>
                <DialogContent
                    DialogOverlayProps={{
                        blurBackground: false,
                    }}
                    maxWidth='sm'>
                    <DialogTitle>
                        Tag Results to {title} {isLabel.toString()}
                    </DialogTitle>

                    <DialogDescription>{description}</DialogDescription>

                    <Select onValueChange={handleValueChange}>
                        <SelectTrigger className='w-60'>
                            <SelectValue placeholder={`Select ${title}`} />
                        </SelectTrigger>
                        <SelectPortal>
                            <SelectContent>
                                {!isLabel &&
                                    zones?.map((zone) => {
                                        return (
                                            <SelectItem key={zone.id} value={zone.id.toString()}>
                                                {zone.name}
                                            </SelectItem>
                                        );
                                    })}
                                {isLabel &&
                                    labels?.map((label) => {
                                        return (
                                            <SelectItem key={label.id} value={label.id.toString()}>
                                                {label.name}
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

                        <Button disabled={continueDisabled} onClick={onContinue}>
                            Continue
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default TagToZoneDialog;
