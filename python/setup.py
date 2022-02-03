import setuptools

from labe import __version__

with open("README.md", "r") as fh:
    long_description = fh.read()

    setuptools.setup(
        name="labe",
        version=__version__,
        author="Martin Czygan",
        author_email="martin.czygan@gmail.com",
        description="Citation graph application at SLUB Dresden (LABE)",
        long_description=long_description,
        long_description_content_type="text/markdown",
        url="https://github.com/slub/labe",
        packages=setuptools.find_packages(),
        classifiers=[
            "Programming Language :: Python :: 3",
            "License :: OSI Approved :: MIT License",
            "Operating System :: OS Independent",
        ],
        python_requires=">=3.6",
        entry_points={"console_scripts": ["labe=labe.cli:main"]},
        install_requires=[
            "luigi>=3,<4",
            "pandas>=1.3.5,<2",
            "requests>=2.26,<3",
            "xdg>=5,<6",
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
