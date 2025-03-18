import { ak_call, Waiter } from './ak_call';

export function initializeGlobals(waiter: Waiter) {
    Object.defineProperty(global, 'ak_call', {
        value: ak_call(waiter),
        configurable: false,
        writable: false
    });
}
