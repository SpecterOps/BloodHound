import { Tabs, TabsList, TabsTrigger } from '@bloodhoundenterprise/doodleui';
import { CircularProgress } from '@mui/material';
import { FC, Suspense } from 'react';
import { usePZPathParams } from '../../../hooks';
import { SelectedDetailsTabContent } from './SelectedDetailsTabContent';
import { DetailsTabOption, detailsTabOptions } from './utils';

type SelectedDetailsTabsProps = {
    currentDetailsTab: DetailsTabOption;
    onTabClick: (value: DetailsTabOption) => void;
};

export const SelectedDetailsTabs: FC<SelectedDetailsTabsProps> = ({ currentDetailsTab, onTabClick }) => {
    const { memberId, ruleId, tagTypeDisplay } = usePZPathParams();
    return (
        <>
            <Tabs
                defaultValue={currentDetailsTab}
                value={currentDetailsTab}
                className='w-full mb-4'
                onValueChange={(value) => {
                    onTabClick(value as DetailsTabOption);
                }}>
                <TabsList className='w-full flex justify-start'>
                    <TabsTrigger value={detailsTabOptions[1]}>{tagTypeDisplay}</TabsTrigger>
                    <TabsTrigger disabled={!ruleId} value={detailsTabOptions[2]}>
                        Rule
                    </TabsTrigger>
                    <TabsTrigger disabled={!memberId} value={detailsTabOptions[3]}>
                        Object
                    </TabsTrigger>
                </TabsList>
            </Tabs>
            <Suspense
                fallback={
                    <div className='absolute inset-0 flex items-center justify-center'>
                        <CircularProgress color='primary' size={80} />
                    </div>
                }>
                <SelectedDetailsTabContent currentDetailsTab={currentDetailsTab} />
            </Suspense>
        </>
    );
};
