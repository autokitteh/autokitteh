from conftest import tests_dir
import tomllib

runner_project_file = tests_dir / "../pyproject.toml"
sdk_project_file = tests_dir / "../../py-sdk/pyproject.toml"


def test_sdk_version():
    """Test that runner SDK major version is same as py-sdk"""
    with open(runner_project_file, "rb") as fp:
        runner_prj = tomllib.load(fp)

    for dep in runner_prj["project"]["dependencies"]:
        if dep.startswith("autokitteh"):
            ak_dep = dep
            break
    else:
        assert False, f"can't find autokitteh dependency in {runner_project_file}"

    # autokitteh[all] ~= 0.4
    runner_ver = ak_dep.split("=")[-1]
    runner_ver = [int(v) for v in runner_ver.split(".")]

    with open(sdk_project_file, "rb") as fp:
        sdk_prj = tomllib.load(fp)

    # 0.4.0
    sdk_ver = sdk_prj["project"]["version"]
    sdk_ver = [int(v) for v in sdk_ver.split(".")]

    # Check up to minor version
    assert runner_ver[:2] == sdk_ver[:2]
