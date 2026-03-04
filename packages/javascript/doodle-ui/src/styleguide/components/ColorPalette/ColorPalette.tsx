import presets from '../../../tailwind/preset';

const colors = presets.theme.extend.colors;

interface SwatchObject {
    objKey: string;
}

const Swatch = (props: SwatchObject) => {
    const { objKey } = props;
    return (
        <div className='border mb-4 flex flex-col items-center p-4'>
            <div className={`w-16 h-16 mb-1 border bg-${objKey}`}></div>
            <p>{objKey}</p>
        </div>
    );
};

interface ElevationProps {
    i: number;
}
const ElevationTile = (props: ElevationProps) => {
    const { i } = props;
    return (
        <div className={`top-0 left-0 flex absolute top-${i * 8} left-${i * 8}`}>
            <div className={`w-32 h-32 bg-neutral-${i + 1} border mr-4 pt-4`}></div>
            <div className='text-body-2'>bg-neutral-{i}</div>
        </div>
    );
};

const ColorPalette = () => {
    return (
        <div className='mb-8'>
            <h2 className='text-headline-2 text-transparent bg-clip-text font-extrabold bg-gradient-to-r from-primary to-error inline-block'>
                Brand
            </h2>
            <div className='grid w-full grid-cols-3 gap-4'>
                {Object.keys(colors).map((key, index) => {
                    if (index < 7) {
                        return <Swatch key={key} objKey={key} />;
                    }
                })}
            </div>

            <h2 className='text-headline-2 text-transparent bg-clip-text font-extrabold bg-gradient-to-r from-primary to-error inline-block'>
                Text
            </h2>

            <div className='grid w-full grid-cols-3 gap-4 mb-4'>
                {Object.keys(colors).map((key, index) => {
                    if (index > 6 && index < 11) {
                        return <Swatch key={key} objKey={key} />;
                    }
                })}
            </div>

            <h2 className='text-headline-2 text-transparent bg-clip-text font-extrabold bg-gradient-to-r from-primary to-error inline-block'>
                Elevations
            </h2>

            <div className='flex mt-4'>
                <div className='w-full mb-16 h-64'>
                    <div className='relative'>
                        <div className='absolute top-0 left-0 flex'>
                            <div className='w-32 h-32 bg-neutral-1 border absolute top-0 left-0'></div>
                            <div>bg-neutral-1</div>
                        </div>
                        {[...Array(5)].map((_, i) => {
                            return <ElevationTile key={i} i={i} />;
                        })}
                    </div>
                </div>
            </div>

            <div className='grid w-full grid-cols-3 gap-4 mb-12'>
                {Object.keys(colors).map((key, index) => {
                    if (index > 9 && index < 15) {
                        return <Swatch key={key} objKey={key} />;
                    }
                })}
            </div>

            <div className='grid w-full grid-cols-3 gap-4 mb-12'>
                {Object.keys(colors).map((key, index) => {
                    if (index > 14 && index < 20) {
                        return <Swatch key={key} objKey={key} />;
                    }
                })}
            </div>
            <div className='grid w-full grid-cols-3 gap-4 mb-12'>
                {Object.keys(colors).map((key, index) => {
                    if (index > 19 && index < 25) {
                        return <Swatch key={key} objKey={key} />;
                    }
                })}
            </div>
        </div>
    );
};

export { ColorPalette };
