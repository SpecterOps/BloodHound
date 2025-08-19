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
import { useTagsQuery } from '../../../hooks';
import { QueryLineItem } from '../../../types';
import { useAppNavigate } from '../../../utils';

type TagToZoneLabelDialogProps = {
    dialogOpen: boolean;
    selectedQuery: QueryLineItem | undefined;
    isLabel: boolean;
    cypherQuery: string;
    setDialogOpen: (isOpen: boolean) => void;
};

const TagToZoneLabelDialog = (props: TagToZoneLabelDialogProps) => {
    const { dialogOpen, selectedQuery, isLabel, cypherQuery, setDialogOpen } = props;
    const navigate = useAppNavigate();

    const AssetGroupTagTypeTier = 1 as const;
    const AssetGroupTagTypeLabel = 2 as const;

    const tiersQuery = useTagsQuery((tag) => tag.type === AssetGroupTagTypeTier);
    const zones = tiersQuery.data;

    const labelsQuery = useTagsQuery((tag) => tag.type === AssetGroupTagTypeLabel);
    const labels = labelsQuery.data;

    const [zone, setZone] = useState('');
    const [label, setLabel] = useState('');

    const handleValueChange = (val: string) => {
        isLabel ? setLabel(val) : setZone(val);
    };

    const stateToPass = cypherQuery ? { query: cypherQuery } : selectedQuery;

    const onContinue = () => {
        if (isLabel) {
            navigate(`/zone-management/save/label/${label}/selector`, { state: stateToPass });
        } else {
            navigate(`/zone-management/save/tier/${zone}/selector`, { state: stateToPass });
        }
    };

    const title = isLabel ? 'Label' : 'Zone';

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
                    <DialogTitle>Tag Results to {title}</DialogTitle>

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

export default TagToZoneLabelDialog;
