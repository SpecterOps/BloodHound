import { KeyboardEvent, KeyboardEventHandler } from 'react';

export function adaptClickHandlerToKeyDown(
    event: KeyboardEvent<HTMLElement>,
    handler: KeyboardEventHandler<HTMLElement>
) {
    if ('key' in event) {
        if (event.key === 'Enter' || event.key === ' ') {
            handler(event);
        }
    }
}

// function adaptClickHandlerToKeyDown(divEventHandler: any) {
//     return (event: KeyboardEvent<HTMLDivElement>) => {
//         if ('key' in event) {
//             if (event.key === 'Enter' || event.key === ' ') {
//                 divEventHandler(event);
//             }
//         } else {
//             divEventHandler(event);
//         }
//     };
// }

// export const AccessibleClickableDiv = ({
//     children,
//     onClickAccessible,
//     ...rest
// }: React.ComponentProps<'div'> & {
//     onClickAccessible: KeyboardEventHandler<HTMLDivElement> | MouseEventHandler<HTMLDivElement>;
// }) => (
//     <div
//         role='button'
//         onKeyDown={adaptClickHandlerToKeyDown(onClickAccessible as KeyboardEventHandler<HTMLDivElement>)}
//         tabIndex={0}
//         {...rest}>
//         {children}
//     </div>
// );
