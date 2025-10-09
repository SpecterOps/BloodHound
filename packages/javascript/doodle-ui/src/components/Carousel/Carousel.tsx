import * as React from 'react';
import { EmblaCarouselType } from 'embla-carousel';
import useEmblaCarousel, { type UseEmblaCarouselType } from 'embla-carousel-react';
import Autoplay from 'embla-carousel-autoplay';

import { cn } from 'components/utils';
import { Button } from 'components/Button';
import { ChevronLeft, ChevronRight } from 'lucide-react';

type CarouselApi = UseEmblaCarouselType[1];
type UseCarouselParameters = Parameters<typeof useEmblaCarousel>;
type CarouselOptions = UseCarouselParameters[0];

type CarouselProps = {
    opts?: CarouselOptions;
    autoplay?: boolean;
    orientation?: 'horizontal' | 'vertical';
    setApi?: (api: CarouselApi) => void;
};

type CarouselContextProps = {
    carouselRef: ReturnType<typeof useEmblaCarousel>[0];
    api: ReturnType<typeof useEmblaCarousel>[1];
    scrollPrev: () => void;
    scrollNext: () => void;
    onDotButtonClick: (index: number) => void;
    selectedIndex: number;
    scrollSnaps: number[];
    canScrollPrev: boolean;
    canScrollNext: boolean;
} & CarouselProps;

const CarouselContext = React.createContext<CarouselContextProps | null>(null);

function useCarousel() {
    const context = React.useContext(CarouselContext);

    if (!context) {
        throw new Error('useCarousel must be used within a <Carousel />');
    }

    return context;
}

const Carousel = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & CarouselProps>(
    ({ orientation = 'horizontal', autoplay = false, opts, setApi, className, children, ...props }, ref) => {
        const [carouselRef, api] = useEmblaCarousel(
            {
                ...opts,
                axis: orientation === 'horizontal' ? 'x' : 'y',
            },
            [...(autoplay ? [Autoplay()] : [])]
        );
        const [canScrollPrev, setCanScrollPrev] = React.useState(false);
        const [canScrollNext, setCanScrollNext] = React.useState(false);
        const [selectedIndex, setSelectedIndex] = React.useState(0);
        const [scrollSnaps, setScrollSnaps] = React.useState<number[]>([]);

        const onDotButtonClick = React.useCallback(
            (index: number) => {
                if (!api) return;
                api.scrollTo(index);
            },
            [api]
        );

        const onInit = React.useCallback((api: EmblaCarouselType) => {
            setScrollSnaps(api.scrollSnapList());
        }, []);

        const onSelect = React.useCallback((api: CarouselApi) => {
            if (!api) {
                return;
            }
            setCanScrollPrev(api.canScrollPrev());
            setCanScrollNext(api.canScrollNext());
            setSelectedIndex(api.selectedScrollSnap());
        }, []);

        const scrollPrev = React.useCallback(() => {
            api?.scrollPrev();
        }, [api]);

        const scrollNext = React.useCallback(() => {
            api?.scrollNext();
        }, [api]);

        const handleKeyDown = React.useCallback(
            (event: React.KeyboardEvent<HTMLDivElement>) => {
                if (event.key === 'ArrowLeft') {
                    event.preventDefault();
                    scrollPrev();
                } else if (event.key === 'ArrowRight') {
                    event.preventDefault();
                    scrollNext();
                }
            },
            [scrollPrev, scrollNext]
        );

        React.useEffect(() => {
            if (!api || !setApi) {
                return;
            }

            setApi(api);
        }, [api, setApi]);

        React.useEffect(() => {
            if (!api) {
                return;
            }
            onInit(api);
            onSelect(api);
            api.on('reInit', onInit);
            api.on('reInit', onSelect);
            api.on('select', onSelect);

            return () => {
                api?.off('select', onSelect);
            };
        }, [api, onInit, onSelect]);

        return (
            <CarouselContext.Provider
                value={{
                    carouselRef,
                    api: api,
                    opts,
                    orientation: orientation || (opts?.axis === 'y' ? 'vertical' : 'horizontal'),
                    scrollPrev,
                    scrollNext,
                    selectedIndex,
                    scrollSnaps,
                    canScrollPrev,
                    canScrollNext,
                    onDotButtonClick,
                }}>
                <div
                    ref={ref}
                    onKeyDownCapture={handleKeyDown}
                    className={cn('relative', className)}
                    role='region'
                    aria-roledescription='carousel'
                    {...props}>
                    {children}
                </div>
            </CarouselContext.Provider>
        );
    }
);
Carousel.displayName = 'Carousel';

const CarouselContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
    ({ className, ...props }, ref) => {
        const { carouselRef, orientation } = useCarousel();

        return (
            <div ref={carouselRef} className='overflow-hidden'>
                <div
                    ref={ref}
                    className={cn('flex', orientation === 'horizontal' ? '-ml-4' : '-mt-4 flex-col', className)}
                    {...props}
                />
            </div>
        );
    }
);
CarouselContent.displayName = 'CarouselContent';

const CarouselItem = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
    ({ className, ...props }, ref) => {
        const { orientation } = useCarousel();

        return (
            <div
                ref={ref}
                role='group'
                aria-roledescription='slide'
                className={cn(
                    'min-w-0 shrink-0 grow-0 basis-full',
                    orientation === 'horizontal' ? 'pl-4' : 'pt-4',
                    className
                )}
                {...props}
            />
        );
    }
);
CarouselItem.displayName = 'CarouselItem';

const CarouselPrev = React.forwardRef<HTMLButtonElement, React.ComponentProps<typeof Button>>(({ ...props }, ref) => {
    const { scrollPrev, canScrollPrev } = useCarousel();

    return (
        <Button
            name='prev'
            ref={ref}
            className='px-2'
            variant='text'
            size='small'
            disabled={!canScrollPrev}
            onClick={scrollPrev}
            {...props}>
            <ChevronLeft className='h-3 w-3 text-black dark:text-white' />
        </Button>
    );
});
CarouselPrev.displayName = 'CarouselPrev';

const CarouselDots = React.forwardRef<HTMLButtonElement, React.ComponentProps<typeof Button>>(() => {
    const { selectedIndex, onDotButtonClick, scrollSnaps } = useCarousel();

    return (
        <div className='flex flex-row'>
            {scrollSnaps.map((_, index) => (
                <button
                    name={`dot-${index}`}
                    key={index}
                    className={'h-2.5 w-2.5 mx-1.5 border-x border-y rounded-3xl border-black dark:border-white '.concat(
                        index === selectedIndex ? 'active bg-black dark:bg-white rounded-3xl' : ''
                    )}
                    onClick={() => onDotButtonClick(index)}></button>
            ))}
        </div>
    );
});
CarouselDots.displayName = 'CarouselDots';

const CarouselNext = React.forwardRef<HTMLButtonElement, React.ComponentProps<typeof Button>>(({ ...props }, ref) => {
    const { scrollNext, canScrollNext } = useCarousel();

    return (
        <Button
            name='next'
            ref={ref}
            className='px-2'
            variant='text'
            size='small'
            disabled={!canScrollNext}
            onClick={scrollNext}
            {...props}>
            <ChevronRight className='h-3 w-3 text-black dark:text-white' />
        </Button>
    );
});
CarouselNext.displayName = 'CarouselNext';

export { type CarouselApi, Carousel, CarouselContent, CarouselItem, CarouselPrev, CarouselDots, CarouselNext };
