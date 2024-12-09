from inspect import currentframe
from pathlib import Path


def iter_frames():
    frame = currentframe()
    while frame:
        yield frame
        frame = frame.f_back


def code_loc(code_dir: Path):
    """Find first location of user code, which is the deepest frame."""
    for frame in iter_frames():
        file_name = Path(frame.f_code.co_filename)
        if not file_name.is_relative_to(code_dir):
            continue

        return file_name, frame.f_lineno

    return "", -1


def make_audit_hook(ak_call, code_dir: Path):
    # TODO(ENG-1893): Do we want more events (i.e. fcntl.*)?
    # See full list at https://docs.python.org/3/library/audit_events.html
    def hook(event, args):
        if ak_call.in_activity or event != "open":
            return

        file_name, line_num = code_loc(code_dir)
        if not file_name:
            return

        print(f"WARNING: {file_name}:{line_num}: file operation not inside an activity")

    return hook
