# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html
from pathlib import Path
import sys
import tomllib

sdk_dir = Path(__file__).parent.parent.absolute()
sys.path.insert(0, str(sdk_dir))  # Add the project root to the path

# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information

project = "autokitteh"
copyright = "2024, AutoKitteh"
author = "AutoKitteh"
project_file = sdk_dir / "pyproject.toml"
with project_file.open("rb") as fp:
    data = tomllib.load(fp)
    release = data["project"]["version"]


# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = ["sphinx.ext.napoleon"]

templates_path = ["_templates"]
exclude_patterns = ["_build", "Thumbs.db", ".DS_Store"]


# -- Options for HTML output -------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#options-for-html-output

html_theme = "alabaster"
html_static_path = ["_static"]
