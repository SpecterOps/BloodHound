// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
import type { Meta, StoryObj } from '@storybook/react';

import { Carousel, CarouselContent, CarouselDots, CarouselItem, CarouselNext, CarouselPrev } from './Carousel';

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
