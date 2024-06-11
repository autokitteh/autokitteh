import json
import pickle
from base64 import b64encode, b64decode


class MessageType:
    """Possible message types."""
    callback = 'callback'
    done = 'done'
    log = 'log'
    module = 'module'
    response = 'response'
    run = 'run'
    sleep = 'sleep'
    

class Comm:
    """Comm does communication with ak server, JSON lines over socket."""
    def __init__(self, sock):
        self.sock = sock
        self.rdr = sock.makefile('r')

    def _send(self, message):
        data = json.dumps(message) + '\n'
        self.sock.sendall(data.encode('utf-8'))

    def recv(self, *msg_types):
        data = self.rdr.readline()
        if not data:
            raise ValueError('connection closed')

        message = json.loads(data)
        if (typ := message['type']) not in msg_types:
            typs = ', '.join(msg_types)
            raise ValueError(f'message type: expected one of {typs!r}, got {typ!r}')
        return message

    def _picklize(self, data):
        data = pickle.dumps(data, protocol=0)
        return b64encode(data).decode('utf-8')

    def send_activity(self, fn, args, kw):
        data = (fn, args, kw)
        message = {
            'type': MessageType.callback,
            'payload': {
                'name': fn if isinstance(fn, str) else fn.__name__,
                'args': [repr(a) for a in args],
                'kw': {k: repr(v) for k, v in kw.items()},
                'data': self._picklize(data),
            },
        }
        self._send(message)

    def extract_activity(self, message):
        payload = message['payload']
        data = b64decode(payload['data'])
        payload['data'] = pickle.loads(data)
        return payload

    def send_exported(self, entries):
        message = {
            'type': MessageType.module,
            'payload': {
                'entries': entries,
            }
        }
        self._send(message)

    def send_done(self):
        message = {'type': MessageType.done}
        self._send(message)

    def receive_run(self):
        message = self.recv(MessageType.run)
        return message['payload']

    def send_response(self, value):
        message = {
            'type': MessageType.response,
            'payload': {
                'value': self._picklize(value),
            }
        }
        self._send(message)

    def extract_response(self, message):
        data = message['payload']['value']
        return pickle.loads(b64decode(data))


    def send_log(self, level, message):
        message = {
            'type': MessageType.log,
            'payload': {
                'level': level,
                'message': message,
            },
        }
        self._send(message)

    def send_sleep(self, seconds):
        message = {
            'type': MessageType.sleep,
            'payload': {
                'seconds': seconds,
            },
        }
        self._send(message)
