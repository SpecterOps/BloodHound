import { faCaretDown, faCaretUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { cn, formatPotentiallyUnknownLabel } from '../../utils';
import { MungedTableRowWithId } from './explore-table-utils';

const ExploreTableHeaderCell = ({
    headerKey,
    sortBy,
    sortOrder,
    onClick,
}: {
    headerKey: keyof MungedTableRowWithId;
    sortBy?: keyof MungedTableRowWithId;
    sortOrder?: string;
    onClick: () => void;
}) => {
    return (
        <div
            className='flex items-center p-1 m-0 cursor-pointer h-full hover:bg-neutral-100 dark:hover:bg-neutral-dark-4'
            onClick={onClick}>
            <div>{formatPotentiallyUnknownLabel(String(headerKey))}</div>
            <div className={cn('pl-2', { ['opacity-0']: sortBy !== headerKey })}>
                {!sortOrder && <FontAwesomeIcon icon={faCaretDown} />}
                {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} />}
                {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} />}
            </div>
        </div>
    );
};

export default ExploreTableHeaderCell;
