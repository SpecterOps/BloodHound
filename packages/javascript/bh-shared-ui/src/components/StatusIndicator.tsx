export type StatusType = 'good' | 'bad' | 'pending';

type Props = {
    label?: string;
    pulse?: boolean;
    type: StatusType;
};

const STATUS_COLORS: Record<StatusType, string> = {
    good: 'fill-[#BCD3A8]', // Light Olive Green
    bad: 'fill-[#D9442E]', // Red
    pending: 'fill-[#5CC3AD]', // Aqua Green
};

/**
 * Displays a colored circle and label corresponding to a given status value and type.
 *
 * @example
 * ```tsx
 * const status = <StatusIndicator status={jobStatus} type="job" />
 * ```
 */
export const StatusIndicator: React.FC<Props> = ({ label = '', pulse = false, type }) => {
    const color = STATUS_COLORS[type];
    return (
        <span className='inline-flex items-center'>
            <span className='mr-1.5'>
                <svg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12' fill='none'>
                    <circle cx='6' cy='6' r='6' className={`${color}${pulse ? ' animate-pulse' : ''}`} />
                </svg>
            </span>
            {label && <span>{label}</span>}
        </span>
    );
};
