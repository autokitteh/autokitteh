import json
from socket import socketpair

from ak_runner.comm import Comm, MessageType


def sub(a, b, *, verbose=False):
    if verbose:
        print(f'{a} - {b}')
    return a - b



def test_comm():
    go, py = socketpair()

    # Callback
    comm = Comm(py)
    args, kw = (1, 7), {'verbose': False}
    comm.send_activity(sub, args, kw)
    data = go.recv(2048)
    assert data, 'no data'

    go.sendall(data)
    message = comm.recv(MessageType.callback)
    payload = comm.extract_activity(message)
    assert payload['name'] == sub.__name__
    assert payload['args'] == [str(v) for v in args]
    assert payload['kw'] == {k: str(v) for k, v in kw.items()}
    fn, args, kw = payload['data']
    assert fn == sub
    assert args == args
    assert kw == kw

    # Module
    names = ['a', 'c', 'f']
    comm.send_exported(names)
    data = go.recv(2048)
    assert data, 'no data'
    message = json.loads(data)
    assert message['type'] == MessageType.module
    assert message['payload']['entries'] == names


    # Done
    comm.send_done()
    data = go.recv(2048)
    assert data, 'no data'
    message = json.loads(data)
    assert message['type'] == MessageType.done


