# Generate by gen_py_messages.go, DO NOT EDIT

from dataclasses import dataclass, asdict, field, fields
import json
from base64 import b64encode, b64decode
from abc import ABC, abstractmethod

def encode(value):
    if not isinstance(value, bytes):
        return value

    value = b64encode(value)
    return value.decode('utf-8')


@dataclass
class Message(ABC):
    @abstractmethod
    def type(self) -> str:
        ...

    def as_dict(self):
        return {k: encode(v) for k, v in asdict(self).items()}

    @classmethod
    def from_payload(cls, payload: dict):
        # TODO(?): generate once
        binary_fields = {f.name for f in fields(cls) if f.type is bytes}
        obj = {k: cls.decode(k, v, binary_fields) for k, v in payload.items()}
        return cls(**obj)

    @classmethod
    def decode(cls, key, value, binary_fields):
        if key not in binary_fields:
            return value
        return b64decode(value)


def encode_message(message):
    obj = {
        'type': message.type(),
        'payload': message.as_dict(),
    }
    return json.dumps(obj)


def decode_message(data):
    msg = json.loads(data)
    # dispatch is defined by gen_py_messages.go
    cls = dispatch.get(msg['type'])
    if not cls:
        raise ValueError(f'unknown message type: {msg!r}')

    return cls.from_payload(msg['payload'])
