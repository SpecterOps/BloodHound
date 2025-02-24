import { SearchValue } from '../../../store';

export const useNodeSearch = () => {
    return {
        searchTerm: 'To be implemented',
        selectedItem: { objectid: 'To be implemented' },
        editSourceNode: (edit: string) => console.log(edit),
        selectSourceNode: (selected?: SearchValue) => console.log(selected),
    };
};
