import presets from '../../../tailwind/preset';

const fonts = presets.theme.extend.fontSize;

const FontSizes = () => {
    return (
        <div className='mb-8'>
            {Object.entries(fonts).map(([key, val]) => (
                <div key={key} className={`mb-5`}>
                    <div className={`grid grid-cols-3 leading-none text-${key}`}>
                        <p>{key}</p>
                        <p className='font-bold'>{key}</p>
                        {key !== 'headline-1' && key !== 'headline-2' && <p className='underline'>{key}</p>}
                    </div>
                    <p>
                        <span className='text-eyeline'>{val}</span>
                    </p>
                </div>
            ))}
        </div>
    );
};

export { FontSizes };
