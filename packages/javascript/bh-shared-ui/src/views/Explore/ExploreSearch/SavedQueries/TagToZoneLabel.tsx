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
import { Button, Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { FC, useState } from 'react';
import { AppIcon } from '../../../../components';
import { adaptClickHandlerToKeyDown } from '../../../../utils/adaptClickHandlerToKeyDown';
import { useSavedQueriesContext } from '../../providers';
import TagToZoneLabelDialog from './TagToZoneLabelDialog';

type TagToZoneLabelProps = {
    cypherQuery: string;
};

const TagToZoneLabel: FC<TagToZoneLabelProps> = (props) => {
    const { cypherQuery } = props;

    const { selectedQuery } = useSavedQueriesContext();

    const listItemStyles = 'px-2 py-3 cursor-pointer hover:bg-neutral-4';

    const [tagToZoneOpen, setTagToZoneOpen] = useState(false);
    //Tag to Zone or Label
    const [isLabel, setIsLabel] = useState(false);

    const handleSetOpen = (isOpen: boolean) => {
        setTagToZoneOpen(isOpen);
    };

    const tagToZone = () => {
        setIsLabel(false);
        setTagToZoneOpen(true);
    };

    const tagToLabel = () => {
        setIsLabel(true);
        setTagToZoneOpen(true);
    };

    return (
        <>
            <Popover>
                <PopoverTrigger disabled={!selectedQuery && !cypherQuery}>
                    <Button variant='secondary' asChild size='small'>
                        <div>
                            <span className='mr-2 text-base'>Tag</span>
                            <AppIcon.CaretDown size={10} />
                        </div>
                    </Button>
                </PopoverTrigger>
                <PopoverContent className='p-0 w-28'>
                    <div
                        role='button'
                        tabIndex={0}
                        onKeyDown={adaptClickHandlerToKeyDown(tagToZone)}
                        className={listItemStyles}
                        onClick={tagToZone}>
                        Zone
                    </div>
                    <div
                        role='button'
                        tabIndex={0}
                        onKeyDown={adaptClickHandlerToKeyDown(tagToLabel)}
                        className={listItemStyles}
                        onClick={tagToLabel}>
                        Label
                    </div>
                </PopoverContent>
            </Popover>
            <TagToZoneLabelDialog
                dialogOpen={tagToZoneOpen}
                setDialogOpen={handleSetOpen}
                isLabel={isLabel}
                selectedQuery={selectedQuery}
                cypherQuery={cypherQuery}
            />
        </>
    );
};

export default TagToZoneLabel;
