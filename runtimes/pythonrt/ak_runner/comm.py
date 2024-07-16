import json
import pickle
from base64 import b64decode, b64encode
from traceback import extract_stack


class MessageType:
    """Possible message types."""

    call = "call"
    call_return = "return"
    callback = "callback"
    error = "error"
    done = "done"
    log = "log"
    module = "module"
    response = "response"
    run = "run"
    sleep = "sleep"


class Comm:
    """Comm communicates with ak server.

    The communication protocol is JSON object per line over a socket.
    See "Communication Sequence" section of the README for more details.
    """

    def __init__(self, sock):
        self.sock = sock
        self.rdr = sock.makefile("r")

    def _send(self, message):
        data = json.dumps(message) + "\n"
        self.sock.sendall(data.encode("utf-8"))

    def recv(self, *msg_types):
        data = self.rdr.readline()
        if not data:
            raise ValueError("connection from autokitteh closed")

        message = json.loads(data)
        if (typ := message["type"]) not in msg_types:
            typs = ", ".join(msg_types)
            raise ValueError(f"message type: expected one of {typs!r}, got {typ!r}")
        return message

    def _picklize(self, data):
        data = pickle.dumps(data, protocol=0)
        return b64encode(data).decode("utf-8")

    def send_activity(self, fn, args, kw):
        data = (fn, args, kw)
        message = {
            "type": MessageType.callback,
            "payload": {
                "name": fn if isinstance(fn, str) else fn.__name__,
                "args": [repr(a) for a in args],
                "kw": {k: repr(v) for k, v in kw.items()},
                "data": self._picklize(data),
            },
        }
        self._send(message)

    def extract_activity(self, message):
        payload = message["payload"]
        data = b64decode(payload["data"])
        payload["data"] = pickle.loads(data)
        return payload

    def send_exported(self, entries):
        message = {
            "type": MessageType.module,
            "payload": {
                "entries": entries,
            },
        }
        self._send(message)

    def send_done(self):
        message = {"type": MessageType.done}
        self._send(message)

    def receive_run(self):
        message = self.recv(MessageType.run)
        return message["payload"]

    def send_response(self, value):
        message = {
            "type": MessageType.response,
            "payload": {
                "value": self._picklize(value),
            },
        }
        self._send(message)

    def extract_response(self, message):
        data = message["payload"]["value"]
        return pickle.loads(b64decode(data))

    def send_log(self, level, message):
        message = {
            "type": MessageType.log,
            "payload": {
                "level": level,
                "message": message,
            },
        }
        self._send(message)

    def send_call(self, func_name, args):
        message = {
            "type": MessageType.call,
            "payload": {
                "func_name": func_name,
                "args": args,
            },
        }
        self._send(message)

    def send_error(self, error):
        message = {
            "type": MessageType.error,
            "payload": {
                "error": str(error),
                "traceback": format_traceback(error.__traceback__),
            },
        }
        self._send(message)


def format_traceback(tb):
    """Format traceback to JSONable list."""
    return [frame_dict(f) for f in extract_stack(tb.tb_frame)]


def frame_dict(frame):
    return {
        "file": frame.filename,
        "lineno": frame.lineno,
        "code": frame.line,
        "name": frame.name,
    }
