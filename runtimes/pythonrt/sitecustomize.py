# This is the original sitecustomize.py from distroless + AK additions
# It is used during docker builds
import sys

# install the apport exception handler if available
try:
    import apport_python_hook
except ImportError:
    pass
else:
    apport_python_hook.install()

# We don't want to use PYTHONPATH since it sets site-packages before the standard
# library.
sys.path.append("/usr/lib/python3.11/site-packages")
