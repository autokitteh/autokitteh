from sys import executable
from subprocess import run
import json


def test_help():
    cmd = [executable, '-m', 'ak_runner', '-h']
    out = run(cmd)
    assert out.returncode == 0


cls_code = '''
class A: pass

def fn(): pass
'''

def test_class(tmp_path):
    with open(tmp_path / 'cls.py', 'w') as out:
        out.write(cls_code)

    cmd = [executable, '-m', 'ak_runner', 'inspect', str(tmp_path)]
    out = run(cmd, capture_output=True)
    assert out.returncode == 0
    reply = json.loads(out.stdout)
    expected = [
        {'name': 'A', 'file': 'cls.py', 'line': 2},
        {'name': 'fn', 'file': 'cls.py', 'line': 4},
    ]
    assert reply == expected
