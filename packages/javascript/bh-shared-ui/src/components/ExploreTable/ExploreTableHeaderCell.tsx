import { faCaretDown, faCaretUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { cn, formatPotentiallyUnknownLabel } from '../../utils';

const ExploreTableHeaderCell = ({
    headerKey,
    sortBy,
    sortOrder,
    onClick,
}: {
    headerKey: string;
    sortBy?: string;
    sortOrder?: string;
    onClick: () => void;
}) => {
    return (
        <div className='flex items-center p-1 m-0 cursor-pointer h-full hover:bg-neutral-100' onClick={onClick}>
            <div>{formatPotentiallyUnknownLabel(headerKey)}</div>
            <div className={cn('pl-2', sortBy !== headerKey ? 'opacity-0' : '')}>
                {!sortOrder && <FontAwesomeIcon icon={faCaretDown} />}
                {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} />}
                {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} />}
            </div>
        </div>
    );
};

export default ExploreTableHeaderCell;
