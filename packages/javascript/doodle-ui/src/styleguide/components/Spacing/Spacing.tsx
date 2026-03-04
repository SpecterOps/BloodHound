const twValues = [
    {
        value: 0.5,
        rem: 0.125,
        px: 2,
    },
    {
        value: 1,
        rem: 0.25,
        px: 4,
    },
    {
        value: 1.5,
        rem: 0.375,
        px: 6,
    },
    {
        value: 2,
        rem: 0.5,
        px: 8,
    },
    {
        value: 2.5,
        rem: 0.625,
        px: 10,
    },
    {
        value: 3,
        rem: 0.75,
        px: 12,
    },
    {
        value: 3.5,
        rem: 0.875,
        px: 14,
    },
    {
        value: 4,
        rem: 1,
        px: 16,
    },
    {
        value: 5,
        rem: 1.25,
        px: 20,
    },
    {
        value: 6,
        rem: 1.25,
        px: 24,
    },
    {
        value: 7,
        rem: 1.75,
        px: 28,
    },
    {
        value: 8,
        rem: 2,
        px: 32,
    },
    {
        value: 9,
        rem: 2.25,
        px: 36,
    },
    {
        value: 10,
        rem: 2.5,
        px: 40,
    },
    {
        value: 11,
        rem: 2.75,
        px: 44,
    },
    {
        value: 12,
        rem: 3,
        px: 48,
    },
    {
        value: 14,
        rem: 3.5,
        px: 56,
    },
    {
        value: 16,
        rem: 4,
        px: 64,
    },
    {
        value: 20,
        rem: 5,
        px: 80,
    },
    {
        value: 24,
        rem: 6,
        px: 96,
    },
    {
        value: 28,
        rem: 7,
        px: 112,
    },
    {
        value: 32,
        rem: 8,
        px: 128,
    },
    {
        value: 36,
        rem: 9,
        px: 144,
    },
    {
        value: 40,
        rem: 10,
        px: 160,
    },
    {
        value: 44,
        rem: 11,
        px: 176,
    },
    {
        value: 48,
        rem: 12,
        px: 192,
    },
    {
        value: 52,
        rem: 13,
        px: 128,
    },
    {
        value: 56,
        rem: 14,
        px: 224,
    },
    {
        value: 60,
        rem: 15,
        px: 240,
    },

    {
        value: 64,
        rem: 16,
        px: 256,
    },
    {
        value: 72,
        rem: 18,
        px: 288,
    },
    {
        value: 80,
        rem: 20,
        px: 320,
    },
    {
        value: 96,
        rem: 24,
        px: 384,
    },
];

const Spacing = () => {
    return (
        <div className='mb-8'>
            <h2 className='text-headline-3 mb-2'>Tailwind Values</h2>
            <h3 className='text-caption mb-8'>Prefer multiples of 8</h3>

            {twValues.map((space, i) => {
                return (
                    <div key={i} className='mb-10 border rounded p-4'>
                        <div className={`w-${space.value} bg-neutral-3 rounded h-9 mr-4 mb-2`}></div>
                        <span className='font-bold mr-4'>{`w-${space.value}`}</span>
                        <span className='mr-4'>{`${space.rem}rem`}</span>
                        <span>{`${space.px}px`}</span>
                    </div>
                );
            })}
        </div>
    );
};

export { Spacing };
