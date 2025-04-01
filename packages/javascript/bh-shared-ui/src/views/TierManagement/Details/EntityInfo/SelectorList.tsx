import { Popover, PopoverContent, PopoverTrigger } from '@bloodhoundenterprise/doodleui';
import { useState } from 'react';
import { useQuery } from 'react-query';
import { AppIcon } from '../../../../components/AppIcon';
import { apiClient } from '../../../../utils';
import { itemSkeletons } from '../utils';
import EntityInfoCollapsibleSection from './EntityInfoCollapsibleSection';

type SelectorListProps = {
    selectedTag: number;
    selectedObject: number;
};

const SelectorList: React.FC<SelectorListProps> = ({ selectedTag, selectedObject }) => {
    const [menuOpen, setMenuOpen] = useState<{ [key: number]: boolean }>({});

    const selectorsQuery = useQuery(['asset-group-member-info'], () => {
        return apiClient.getAssetGroupLabelMemberInfo(selectedTag, selectedObject).then((res) => {
            return res.data.data['member'];
        });
    });

    const handleMenuClick = (index: number) => {
        setMenuOpen((prev) => ({
            ...prev,
            [index]: !prev[index], //Toggle only the clicked popover
        }));
    };

    const handleViewClick = () => {};
    const handleEditClick = () => {};
    const handleDeleteClick = () => {};

    if (selectorsQuery.isLoading) {
        return itemSkeletons.map((skeleton, index) => {
            return skeleton('object-selector', index);
        });
    }
    if (selectorsQuery.isError) {
        return (
            <li className='border-y-[1px] border-neutral-light-3 dark:border-neutral-dark-3 relative h-10 pl-2'>
                <span className='text-base'>There was an error fetching this data</span>
            </li>
        );
    }

    if (selectorsQuery.isSuccess) {
        return (
            <EntityInfoCollapsibleSection label='Selectors' count={selectorsQuery.data.length}>
                {selectorsQuery.data.map((selector, index) => {
                    return (
                        <div
                            className={`flex items-center gap-2 p-2 ${index % 2 === 0 ? 'bg-[#E3E7EA] dark:bg-[#272727]' : ''}`}
                            key={index}>
                            <Popover open={!!menuOpen[index]}>
                                <PopoverTrigger asChild>
                                    <button onClick={() => handleMenuClick(index)}>
                                        <AppIcon.VerticalEllipsis />
                                    </button>
                                </PopoverTrigger>
                                <PopoverContent
                                    className='w-80 px-4 py-2 flex flex-col gap-2'
                                    onInteractOutside={() => setMenuOpen({})}
                                    onEscapeKeyDown={() => setMenuOpen({})}>
                                    <div
                                        className='cursor-pointer p-2 hover:bg-[#E3E7EA] hover:dark:bg-[#272727]'
                                        onClick={handleViewClick}>
                                        View
                                    </div>
                                    <div
                                        className='cursor-pointer p-2 hover:bg-[#E3E7EA] hover:dark:bg-[#272727]'
                                        onClick={handleEditClick}>
                                        Edit
                                    </div>
                                    <div
                                        className='cursor-pointer p-2 hover:bg-[#E3E7EA] hover:dark:bg-[#272727]'
                                        onClick={handleDeleteClick}>
                                        Delete
                                    </div>
                                </PopoverContent>
                            </Popover>
                            {selector.name}
                        </div>
                    );
                })}
            </EntityInfoCollapsibleSection>
        );
    }
};

export default SelectorList;
