from pathlib import Path

here = Path(__file__).parent.absolute()


def test_all_modules_documented():
    docs_index = here / "../docs/index.rst"
    with open(docs_index) as fp:
        index_data = fp.read()

    ignored = {
        "__init__.py",
    }

    no_doc = []
    for mod in (here.parent / "autokitteh").glob("*.py"):
        if mod.name in ignored:
            continue

        ref = f".. automodule:: autokitteh.{mod.stem}"
        if ref not in index_data:
            no_doc.append(mod.name)

    if no_doc:
        no_doc = ", ".join(sorted(no_doc))
        assert False, f"missing from docs/index.rst: {no_doc}"
