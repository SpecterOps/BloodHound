import { DEFAULT_MAP, JOB_STATUS_MAP } from './statusMaps';

const MAP_MAP = {
    default: DEFAULT_MAP,
    job: JOB_STATUS_MAP,
};

type Props = {
    status: number;
    type?: 'default' | 'job';
};

/**
 * Displays a colored circle and label corresponding to a given status value and type.
 *
 * @example
 * ```tsx
 * const status = <StatusIndicator status={jobStatus} type="job" />
 * ```
 */
export const StatusIndicator: React.FC<Props> = ({ status, type = 'default' }) => {
    const { color, name } = MAP_MAP[type][status];
    return (
        <span className='inline-flex items-center'>
            <span className='mr-1.5'>
                <svg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12' fill='none'>
                    <circle cx='6' cy='6' r='6' className={color} />
                </svg>
            </span>
            <span>{name}</span>
        </span>
    );
};
