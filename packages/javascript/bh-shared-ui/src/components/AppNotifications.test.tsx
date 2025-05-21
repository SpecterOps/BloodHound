import { useSnackbar } from 'notistack';
import { SNACKBAR_DURATION, SNACKBAR_DURATION_LONG } from '../constants';
import { useNotifications } from '../providers';
import { act, renderHook, screen } from '../test-utils';
const message = 'This is a notification';
const messageKey = 'messageKey';

describe('AppNotifications', () => {
    it('adds a notification - checks values', () => {
        const hook = renderHook(() => useNotifications());

        act(() =>
            hook.result.current.addNotification(message, messageKey, {
                autoHideDuration: SNACKBAR_DURATION,
            })
        );

        const result = hook.result.current.notifications[0];
        expect(result.message).toBe(message);
        expect(result.key).toBe(messageKey);
        expect(result.options.autoHideDuration).toBe(SNACKBAR_DURATION);
    });

    it('dismisses a notification', () => {
        const hook = renderHook(() => useNotifications());

        act(() => hook.result.current.addNotification(message, messageKey));
        expect(hook.result.current.notifications).toHaveLength(1);
        expect(hook.result.current.notifications[0].dismissed).toBeFalsy;

        act(() => hook.result.current.dismissNotification(messageKey));
        expect(hook.result.current.notifications[0].dismissed).toBeTruthy;
    });

    it('removes a notification', () => {
        const hook = renderHook(() => useNotifications());

        act(() => hook.result.current.addNotification(message, messageKey));
        expect(hook.result.current.notifications).toHaveLength(1);

        act(() => hook.result.current.removeNotification(messageKey));
        expect(hook.result.current.notifications).toHaveLength(0);
    });

    it('renders a snackbar notification in the dom and tests autoHideDuration', async () => {
        const snack = renderHook(() => useSnackbar());
        vi.useFakeTimers();
        await act(() =>
            snack.result.current.enqueueSnackbar('test message', {
                autoHideDuration: SNACKBAR_DURATION,
            })
        );
        expect(await screen.findByText('test message')).toBeInTheDocument();
        //adding 1s cushion to the timer to allow for transition timing
        await vi.advanceTimersByTimeAsync(SNACKBAR_DURATION + 1000);
        expect(screen.queryByText('test message')).not.toBeInTheDocument();
    });

    it('renders a snackbar notification in the dom and tests the long autoHideDuration', async () => {
        const snack = renderHook(() => useSnackbar());
        vi.useFakeTimers();
        await act(() =>
            snack.result.current.enqueueSnackbar('test message', {
                autoHideDuration: SNACKBAR_DURATION_LONG,
            })
        );
        expect(await screen.findByText('test message')).toBeInTheDocument();
        //adding 1s cushion to the timer to allow for transition timing
        await vi.advanceTimersByTimeAsync(SNACKBAR_DURATION_LONG + 1000);
        expect(screen.queryByText('test message')).not.toBeInTheDocument();
    });
});
