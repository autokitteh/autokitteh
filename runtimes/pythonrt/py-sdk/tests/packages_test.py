import sys
from pathlib import Path
from shutil import which
from subprocess import run
from os import environ

import pytest

uv_exe = which("uv")
here = Path(__file__).parent.absolute()
sdk_path = (here / "..").absolute()

py_code = """
from autokitteh import packages

{patch}

# click is small, pure Python and has no external depenecies
packages.install('click')
"""

install_cases = [
    # install command, patch
    pytest.param([uv_exe, "venv"], "", id="uv"),
    pytest.param(
        [sys.executable, "-m", "venv"],
        "packages.which = lambda v: None",
        id="pip",
    ),
]


@pytest.mark.parametrize("venv_cmd, patch", install_cases)
def test_install(venv_cmd, patch, tmp_path):
    if "uv" in venv_cmd[0] and not which("uv"):
        pytest.skip("uv not installed")

    venv_path = tmp_path / "venv"
    venv_cmd += [str(venv_path)]
    out = run(venv_cmd)
    assert out.returncode == 0

    py_script = tmp_path / "main.py"
    with open(py_script, "w") as fp:
        fp.write(py_code.format(patch=patch))

    # Make sure the script can import 'autokitteh'
    env = environ.copy()
    pypath = env.get("PYTHONPATH")
    env["PYTHONPATH"] = f"{sdk_path}:{pypath}"

    venv_py = venv_path / "bin/python"

    # If ak doesn't use uv, this will fail since there's no pip in venv created by uv
    out = run([str(venv_py), str(py_script)], env=env)
    assert out.returncode == 0
