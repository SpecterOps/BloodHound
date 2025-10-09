import type { Meta, StoryObj } from '@storybook/react';

import { Carousel, CarouselContent, CarouselItem, CarouselPrev, CarouselDots, CarouselNext } from './Carousel';

/**
 * A carousel with motion and swipe built using Embla.
 */

const meta = {
    title: 'components/Carousel',
    component: Carousel,
    tags: ['autodocs'],
    argTypes: {},
    args: {
        className: 'w-full max-w-xs',
    },
    parameters: {
        layout: 'centered',
    },
} satisfies Meta<typeof Carousel>;

export default meta;

type Story = StoryObj<typeof meta>;

/**
 * Test data for the story
 */

const TestSlide: React.FC<{ index: number }> = ({ index }) => {
    return (
        <div className='p-1 border-2'>
            <div>
                <div className='flex aspect-square items-center justify-center p-6'>
                    <span className='text-4xl font-semibold text-black dark:text-white'>{index + 1}</span>
                </div>
            </div>
        </div>
    );
};

const TestSlides: React.FC<{ index: number }>[] = [TestSlide, TestSlide, TestSlide];

/**
 * The default form of the carousel.
 */

export const Default: Story = {
    render: () => {
        return (
            <Carousel
                opts={{
                    align: 'center',
                }}
                className='w-full max-w-sm'>
                <CarouselContent>
                    {TestSlides.map((SlideContent, index) => (
                        <CarouselItem key={index}>
                            <SlideContent index={index} />
                        </CarouselItem>
                    ))}
                </CarouselContent>
                <div className='flex justify-center items-center my-1'>
                    <CarouselPrev />
                    <CarouselDots />
                    <CarouselNext />
                </div>
            </Carousel>
        );
    },
};

/**
 * Carousel with autoplay enabled.
 */

export const Autoplay: Story = {
    render: () => {
        return (
            <Carousel
                opts={{
                    align: 'center',
                    loop: true,
                }}
                autoplay={true}
                className='w-full max-w-sm'>
                <CarouselContent>
                    {TestSlides.map((SlideContent, index) => (
                        <CarouselItem key={index}>
                            <SlideContent index={index} />
                        </CarouselItem>
                    ))}
                </CarouselContent>
                <div className='flex justify-center items-center my-1'>
                    <CarouselPrev />
                    <CarouselDots />
                    <CarouselNext />
                </div>
            </Carousel>
        );
    },
};
