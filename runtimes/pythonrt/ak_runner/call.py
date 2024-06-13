import autokitteh

from .comm import Comm, MessageType
from . import log
from time import sleep


def is_marked_activity(fn):
    """Return true if function is marked as an activity."""
    return getattr(fn, autokitteh.ACTIVITY_ATTR, False)


class AKCall:
    """Callable wrapping functions with activities."""
    def __init__(self, comm: Comm):
        self.in_activity = False
        self.comm = comm
        self.module = None

    def is_module_func(self, fn):
        return fn.__module__ == self.module.__name__

    def should_run_as_activity(self, fn):
        if self.in_activity:
            return False

        if is_marked_activity(fn):
            return True

        if fn.__module__ == 'builtins':
            return False
        
        if self.is_module_func(fn):
            return False

        return True

    def __call__(self, func, *args, **kw):
        if not self.should_run_as_activity(func):
            log.info(
                'calling %s (args=%r, kw=%r) directly (in_activity=%s)', 
                func.__name__, args, kw, self.in_activity)
            return func(*args, **kw)

        log.info('ACTION: calling %s via activity (args=%r, kw=%r)', func.__name__, args, kw)
        self.in_activity = True
        try:
            if func is sleep:
                self.comm.send_sleep(*args, **kw)
                self.comm.recv(MessageType.sleep)
                return

            if self.is_module_func(func):
                # Pickle can't handle function from our loaded module
                func = func.__name__
            self.comm.send_activity(func, args, kw)
            message = self.comm.recv(MessageType.callback, MessageType.response)
            
            if message['type'] == MessageType.callback:
                payload = self.comm.extract_activity(message)
                fn, args, kw = payload['data']
                if isinstance(fn, str):
                    fn = getattr(self.module, fn, None)
                    if fn is None:
                        mod_name = self.module.__name__
                        raise ValueError(f'function {fn} not found in {mod_name}')
                value = fn(*args, **kw)
                self.comm.send_response(value)
                message = self.comm.recv(MessageType.response)

            # Reply message, either from current call or playback
            return self.comm.extract_response(message)
        finally:
            self.in_activity = False
