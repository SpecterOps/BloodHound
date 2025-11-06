import { KeyboardEvent, KeyboardEventHandler } from 'react';

export function adaptClickHandlerToKeyDown(handler: KeyboardEventHandler<HTMLElement>) {
    return (event: KeyboardEvent<HTMLElement>) => {
        if ('key' in event) {
            if (event.key === 'Enter' || event.key === ' ') {
                handler(event);
            }
        }
    };
}
