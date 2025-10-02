import { Input } from '@bloodhoundenterprise/doodleui';
import { atom, useAtom } from 'jotai';

export const edgeKindAtom = atom('');

export const GraphEdgeKind = () => {
    const [graphEdgeKind, setGraphEdgeKind] = useAtom(edgeKindAtom);

    return (
        <Input
            className='w-24'
            placeholder='edge type'
            type='text'
            value={graphEdgeKind}
            onInput={(event) => setGraphEdgeKind(event.currentTarget.value)}
        />
    );
};
