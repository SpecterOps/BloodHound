import { KeyboardEvent, KeyboardEventHandler, MouseEvent, MouseEventHandler } from 'react';

export function flexibleKeyboardOrClickHandler(
    event: KeyboardEvent<HTMLElement>,
    handler: KeyboardEventHandler<HTMLElement>
): void;
export function flexibleKeyboardOrClickHandler(
    event: MouseEvent<HTMLElement>,
    handler: MouseEventHandler<HTMLElement>
): void;

export function flexibleKeyboardOrClickHandler(event: any, divEventHandler: any) {
    if ('key' in event) {
        if (event.key === 'Enter' || event.key === ' ') {
            divEventHandler(event as KeyboardEventHandler<HTMLDivElement>);
        }
    } else {
        divEventHandler(event);
    }
}

// function flexibleKeyboardOrClickHandler(divEventHandler: any) {
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
//         onKeyDown={flexibleKeyboardOrClickHandler(onClickAccessible as KeyboardEventHandler<HTMLDivElement>)}
//         tabIndex={0}
//         {...rest}>
//         {children}
//     </div>
// );
