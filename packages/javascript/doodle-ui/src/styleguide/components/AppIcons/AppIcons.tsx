import { Card, CardContent } from '../../../components/Card';
import * as IconOptions from './components';

const componentNames = Object.keys(IconOptions);

const nonLogos = componentNames.filter((x) => !x.includes('Full'));

export type AppIconOptions = keyof typeof IconOptions;

export const AppIcon = IconOptions;

const AppIcons = () => {
    return (
        <div className='mb-8'>
            <div className='mb-4'>
                <div className='grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-8 mb-8'>
                    <div className='flex flex-col items-center justify-center'>
                        <Card className='flex items-center justify-center h-32  mb-2'>
                            <CardContent className='px-6'>
                                <AppIcon.BHCELogoFull size={300} className='text-base' />
                            </CardContent>
                        </Card>
                        <p>BHCELogoFull</p>
                    </div>
                    <div className='flex flex-col items-center justify-center'>
                        <Card className='flex items-center justify-center h-32  mb-2'>
                            <CardContent className='px-6'>
                                <AppIcon.BHELogoFull size={300} className='text-primary' />
                            </CardContent>
                        </Card>
                        <p>BHELogoFull</p>
                    </div>
                </div>
            </div>
            <div className='grid grid-cols-3 md:grid-cols-4 gap-x-4 gap-y-8 mb-4'>
                {nonLogos.map((x: string) => {
                    const MyComponent = AppIcon[x as keyof typeof AppIcon];
                    return (
                        <div key={x} className='flex flex-col items-center justify-center'>
                            <Card className='flex items-center justify-center h-24 w-24 mb-2'>
                                <CardContent>
                                    <MyComponent key={x} size={32} />
                                </CardContent>
                            </Card>
                            <p>{MyComponent.name}</p>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export { AppIcons };
