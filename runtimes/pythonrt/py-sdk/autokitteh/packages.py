import re
import sys
from subprocess import run


def _package_name(specifier):
    """Extract package name from package specifier.

    >>> _package_name('requests')
    'requests'
    >>> _package_name('requests==2.32.2')
    'requests'
    >>> _package_name('requests ~= 2.32')
    'requests'
    """

    match = re.match(r"\w+", specifier)
    if not match:
        raise ValueError(f"bad specifier: {specifier}")

    return match.group()


def _install_package(specifier, import_name):
    try:
        __import__(import_name)
        return
    except ImportError:
        pass

    out = run([sys.executable, "-m", "pip", "install", specifier])
    if out.returncode != 0:
        raise RuntimeError(f"can't install {specifier!r}")

    try:
        __import__(import_name)
    except ImportError:
        raise RuntimeError(
            f"can't import {import_name!r} after installing {specifier!r}"
        )


def install(*packages):
    """Install Python packages using pip.

    A package can be either a package requirement specifier
    (see https://pip.pypa.io/en/stable/reference/requirement-specifiers/)
    or a tuple of (package specifier, import name) in case the import name differs from
    the package name (e.g. Package `pillow` import imported as `PIL`).

    Examples:
    >>> install('requests', 'numpy')
    >>> install('requests ~= 2.32', 'numpy == 2.0.0')
    >>> install(['pillow ~= 10.4', 'PIL'])
    """
    for package in packages:
        if isinstance(package, str):
            specifier = package
            import_name = _package_name(specifier)
        elif isinstance(package, (tuple, list)):
            if len(package) != 2:
                raise ValueError(f"length should be 2: {package}")
            specifier, import_name = package
        else:
            raise TypeError(f"bad package: {package}")

        _install_package(specifier, import_name)
