import setuptools

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setuptools.setup(
    name="autokitteh",
    version="0.0.23",
    author="Itay Donanhirsh",
    author_email="itay@softkitteh.com",
    description="AutoKitteh SDK",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://gitlab.com/softkitteh/autokitteh",
    project_urls={
    },
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    package_dir={"": "src"},
    packages=setuptools.find_namespace_packages(where="src") + setuptools.find_packages(where="src"),
    python_requires=">=3.6",
    zip_safe=False,
    package_data={"autokitteh": ["py.typed"]},
    install_requires=[
        "click",
        "grpc-stubs",
        "grpcio",
        "protobuf",
        "protobuf3",
    ],
)
