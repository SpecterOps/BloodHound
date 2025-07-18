import { faCancel, faCheck } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { EntityField, format } from '../../utils';
import NodeIcon from '../NodeIcon';

const FALLBACK_STRING = '--';

const ExploreTableDataCell = ({ value, columnKey }: { value: EntityField['value']; columnKey: string }) => {
    if (columnKey === 'kind') {
        return (
            <div className='flex justify-center explore-table-cell-icon'>
                <NodeIcon nodeType={value?.toString() || ''} />
            </div>
        );
    }
    if (typeof value === 'boolean' || value === undefined || value === null) {
        return (
            <div className='flex justify-center items-center explore-table-cell-icon pb-1 pt-1'>
                <FontAwesomeIcon
                    icon={value ? faCheck : faCancel}
                    color={value ? 'green' : 'lightgray'}
                    className='scale-125'
                />
            </div>
        );
    }

    const stringyKey = columnKey?.toString();

    return format({ keyprop: stringyKey, value, label: stringyKey }) || FALLBACK_STRING;
};

export default ExploreTableDataCell;
