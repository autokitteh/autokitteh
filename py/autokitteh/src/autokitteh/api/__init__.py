from typing import NewType

import autokitteh.proto.pluginsprovidersvc.svc_pb2 as svc_pb
import autokitteh.proto.pluginsprovidersvc.svc_pb2_grpc as svc_pb_grpc

from autokitteh.proto.plugin.desc_pb2 import PluginDesc, PluginMemberDesc
from autokitteh.proto.program.error_pb2 import Error

from .account import AccountName
from .client import Client
from .eventsrc import EventSourceID, EventSourceName
from .exceptions import AutoKittehException
from .plugins import PluginID, PluginName
from .project import EventSourceProjectBinding
from .values import Value, Func, FuncToValue


EventID = NewType('EventID', str)
ProjectID = NewType('ProjectID', str)


__all__ = [
    'AccountName',
    'AutoKittehException',
    'Client',
    'Error',
    'EventSourceID',
    'EventSourceName',
    'EventSourceProjectBinding',
    'Func',
    'FuncToValue',
    'PluginDesc',
    'PluginID',
    'PluginMemberDesc',
    'PluginName',
    'Value',
    'unwrapped_args',
]
