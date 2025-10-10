import { faArrowTrendUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { Meta, StoryObj } from '@storybook/react';
import { RadialGauge } from '../RadialGauge';
import { ComboBadge } from './ComboBadge';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
    title: 'Components/ComboBadge',
    component: ComboBadge,
    parameters: {
        layout: 'centered',
    },
    // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
    tags: ['autodocs'],
    // More on argTypes: https://storybook.js.org/docs/api/argtypes
    argTypes: {
        label: {
            type: 'string',
            control: 'text',
        },
        adornment: {
            type: 'string',
            control: 'text',
        },
    },
    args: {},
} satisfies Meta<typeof ComboBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SimpleImpl: Story = {
    args: {
        label: '2 📈',
        adornment: '2+',
        type: 'slideLeft',
        ariaLabel: 'x value increased by 2 points',
    },
};

export const ExtendedAdornment: Story = {
    args: {
        label: '2 📈',
        adornment: 'An unreasonably long adornment',
        type: 'slideLeft',
        ariaLabel: 'x value increased by 2 points',
    },
};

export const SlideRight: Story = {
    args: {
        label: '2 📈',
        adornment: '2+',
        type: 'slideRight',
        ariaLabel: 'x value increased by 2 points',
    },
};

export const NoAdornment: Story = {
    args: {
        label: '2 📈',
        type: 'slideRight',
        ariaLabel: 'something',
    },
};

export const InlineAdornmentFigmaSpec: Story = {
    args: {
        label: (
            <>
                2% <FontAwesomeIcon icon={faArrowTrendUp} className='ml-1' />
            </>
        ),
        adornment: '2+',
        type: 'inlineSlideLeft',
        ariaLabel: 'x value increased by 2 points',
        className: 'ml-2',
    },

    render: (props) => {
        return (
            <div className='flex justify-center items-center'>
                <RadialGauge value={50} color='primary' /> 8.1K Exposed Principles
                <ComboBadge {...props} />
            </div>
        );
    },
};
