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
import {
    AssetGroupTag,
    AssetGroupTagTypeLabel,
    AssetGroupTagTypeOwned,
    AssetGroupTagTypeZone,
} from 'js-client-library';
import { useTagsQuery } from '../../../../hooks';
import { labelsPath, privilegeZonesPath, rulesPath, savePath, zonesPath } from '../../../../routes';
import { QueryLineItem } from '../../../../types';
import { useAppNavigate } from '../../../../utils';

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
    const tagsQuery = useTagsQuery();
    const isLabelTagType = (tag: AssetGroupTag) =>
        tag.type === AssetGroupTagTypeLabel || tag.type === AssetGroupTagTypeOwned;
    const isZoneTagType = (tag: AssetGroupTag) => tag.type === AssetGroupTagTypeZone;

    const typeMatcher = isLabel ? isLabelTagType : isZoneTagType;
    const zoneLabelList = tagsQuery.data?.filter(typeMatcher);

    const [zoneId, setZoneId] = useState('');
    const [labelId, setLabelId] = useState('');
    const continueDisabled = (isLabel && !labelId) || (!isLabel && !zoneId);
    const handleValueChange = (val: string) => {
        if (isLabel) {
            setLabelId(val);
            setZoneId('');
        } else {
            setZoneId(val);
            setLabelId('');
        }
    };
    const title = isLabel ? 'Label' : 'Zone';
    const stateToPass = cypherQuery ? { query: cypherQuery } : selectedQuery;

    const onContinue = () => {
        if (isLabel) {
            navigate(`/${privilegeZonesPath}/${labelsPath}/${labelId}/${rulesPath}/${savePath}`, {
                state: stateToPass,
            });
        } else {
            navigate(`/${privilegeZonesPath}/${zonesPath}/${zoneId}/${rulesPath}/${savePath}`, {
                state: stateToPass,
            });
        }
    };

    const description = `Pick a ${title} to create a new rule. All assets returned by the query will be added to your rule.`;

    return (
        <Dialog
            open={dialogOpen}
            onOpenChange={(isOpen) => {
                if (!isOpen) {
                    setDialogOpen(false);
                }
            }}>
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
                                {zoneLabelList?.map((item) => {
                                    return (
                                        <SelectItem key={item.id} value={item.id.toString()}>
                                            {item.name}
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
