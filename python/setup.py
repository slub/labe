import setuptools

from labe import __version__

with open("README.md", "r") as fh:
    long_description = fh.read()

    setuptools.setup(
        name="labe",
        version=__version__,
        author="Martin Czygan",
        author_email="martin.czygan@gmail.com",
        description="Reference data munging tasks and utilities",
        long_description=long_description,
        long_description_content_type="text/markdown",
        url="https://gitlab.com/miku/labe",
        packages=setuptools.find_packages(),
        classifiers=[
            "Programming Language :: Python :: 3",
            "License :: OSI Approved :: MIT License",
            "Operating System :: OS Independent",
        ],
        python_requires=">=3.6",
        entry_points={"console_scripts": ["labectl=labe.cli:main"]},
        install_requires=[
            "luigi",
            "requests",
            "xdg",
            "dynaconf[ini]",
            "marcx",
            "orjson",
        ],
        extras_require={
            "dev": [
                "ipython",
                "isort",
                "mypy",
                "pylint",
                "pytest",
                "pytest-cov",
                "shiv",
                "twine",
                "yapf",
            ],
        },
    )
