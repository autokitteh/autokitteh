from autokitteh_pb.notifications.v1 import notification_pb2 as _notification_pb2
from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SendRequest(_message.Message):
    __slots__ = ["notification"]
    NOTIFICATION_FIELD_NUMBER: _ClassVar[int]
    notification: _notification_pb2.Notification
    def __init__(self, notification: _Optional[_Union[_notification_pb2.Notification, _Mapping]] = ...) -> None: ...

class SendResponse(_message.Message):
    __slots__ = ["notification_id"]
    NOTIFICATION_ID_FIELD_NUMBER: _ClassVar[int]
    notification_id: str
    def __init__(self, notification_id: _Optional[str] = ...) -> None: ...

class ListRequest(_message.Message):
    __slots__ = ["recipient_id", "type", "unread_only", "count_only", "page_size"]
    RECIPIENT_ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    UNREAD_ONLY_FIELD_NUMBER: _ClassVar[int]
    COUNT_ONLY_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    recipient_id: str
    type: str
    unread_only: bool
    count_only: bool
    page_size: int
    def __init__(self, recipient_id: _Optional[str] = ..., type: _Optional[str] = ..., unread_only: bool = ..., count_only: bool = ..., page_size: _Optional[int] = ...) -> None: ...

class ListResponse(_message.Message):
    __slots__ = ["notifications"]
    NOTIFICATIONS_FIELD_NUMBER: _ClassVar[int]
    notifications: _containers.RepeatedCompositeFieldContainer[_notification_pb2.Notification]
    def __init__(self, notifications: _Optional[_Iterable[_Union[_notification_pb2.Notification, _Mapping]]] = ...) -> None: ...

class MarkAsReadRequest(_message.Message):
    __slots__ = ["notification_id"]
    NOTIFICATION_ID_FIELD_NUMBER: _ClassVar[int]
    notification_id: str
    def __init__(self, notification_id: _Optional[str] = ...) -> None: ...

class MarkAsReadResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
