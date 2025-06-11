import { MouseCaptor } from 'sigma';
import { Coordinates, MouseCoords, WheelCoords } from 'sigma/types';

/*
 * This code in this file is copied from the sigma source code, deprecated-v2 branch:
 * https://github.com/jacomyal/sigma.js/blob/deprecated-v2/src/core/captors/mouse.ts
 *
 * Tragically, it seems that changing the ZOOMING_RATIO is the correct answer for
 * "How do we slow down zoom on scroll wheel?" This const is a private immutable quantity
 * living in sigma and is not parameterized, so we have no way to pass it from outside.
 * All this copy paste is to lower the ZOOMING_RATIO from 1.7 to 1.4. and override
 * the default wheel handler with a function using this value. Is there a better way?
 */

const MOUSE_ZOOM_DURATION = 250;
const ZOOMING_RATIO = 1.3;

function getPositionFromSigma(e: MouseEvent | Touch, dom: HTMLElement): Coordinates {
    const bbox = dom.getBoundingClientRect();

    return {
        x: e.clientX - bbox.left,
        y: e.clientY - bbox.top,
    };
}

function getMouseCoordsFromSigma(e: MouseEvent, dom: HTMLElement): MouseCoords {
    const res: MouseCoords = {
        ...getPositionFromSigma(e, dom),
        sigmaDefaultPrevented: false,
        preventSigmaDefault(): void {
            res.sigmaDefaultPrevented = true;
        },
        original: e,
    };

    return res;
}
function getWheelCoordsFromSigma(e: WheelEvent, dom: HTMLElement): WheelCoords {
    return {
        ...getMouseCoordsFromSigma(e, dom),
        delta: getWheelDeltaFromSigma(e),
    };
}

function getWheelDeltaFromSigma(e: WheelEvent): number {
    // TODO: check those ratios again to ensure a clean Chrome/Firefox compat
    if (typeof e.deltaY !== 'undefined') return (e.deltaY * -3) / 360;

    if (typeof e.detail !== 'undefined') return e.detail / -9;

    throw new Error('Captor: could not extract delta from event.');
}

export default function handleWheelFromSigma(this: MouseCaptor, e: WheelEvent): void {
    if (!this.enabled) return;

    e.preventDefault();
    e.stopPropagation();

    const delta = getWheelDeltaFromSigma(e);

    if (!delta) return;

    const wheelCoords = getWheelCoordsFromSigma(e, this.container);
    this.emit('wheel', wheelCoords);

    if (wheelCoords.sigmaDefaultPrevented) return;

    // Default behavior
    const ratioDiff = delta > 0 ? 1 / ZOOMING_RATIO : ZOOMING_RATIO;
    const camera = this.renderer.getCamera();
    const newRatio = camera.getBoundedRatio(camera.getState().ratio * ratioDiff);
    const wheelDirection = delta > 0 ? 1 : -1;
    const now = Date.now();

    // Cancel events that are too close too each other and in the same direction:
    if (
        this.currentWheelDirection === wheelDirection &&
        this.lastWheelTriggerTime &&
        now - this.lastWheelTriggerTime < MOUSE_ZOOM_DURATION / 5
    ) {
        return;
    }

    camera.animate(
        this.renderer.getViewportZoomedState(getPositionFromSigma(e, this.container), newRatio),
        {
            easing: 'quadraticOut',
            duration: MOUSE_ZOOM_DURATION,
        },
        () => {
            this.currentWheelDirection = 0;
        }
    );

    this.currentWheelDirection = wheelDirection;
    this.lastWheelTriggerTime = now;
}
