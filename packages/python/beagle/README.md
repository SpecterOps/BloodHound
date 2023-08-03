# Beagle

Beagle is the BloodHound project cluster build, test and packaging automation framework. The intent of this project is
not to replace tooling like `podman` or `bazel`. The goal of this project is to provide a more ergonomic and automatic
orchestration layer of aforementioned tooling.

Any features added to this framework must first be vetted by identifying if one of the orchestrated tools can cover the
required use-case. If so, beagle should orchestrate its functionality through the applicable tool.

## Usage

```bash
python3 main.py -h 
```

```text
usage: beagle [-h] [-c] [-a] [-v] [-i] {test}

Beagle: the BloodHound build, test and packaging automation framework.

positional arguments:
  {test}             action for beagle to pursue

options:
  -h, --help         show this help message and exit
  -c, --ci           informs beagle that it is running in a CI container
  -a, --all-tests    run all tests regardless of staged changes
  -v, --verbose      enable extra output
  -i, --integration  enable integration tests
```

## Development

Engineers are encouraged to install the following tools in a virtual environment:

```bash
python3 -m venv ~/beagle_venv

~/beagle_venv/bin/pip install --upgrade pip black mypy
```

### Formatting Source

```bash
# From bhce source root
cd packages/python/beagle

~/beagle_venv/bin/black ./ 
```

### Vetting Python Types

```bash
# From bhce source root
cd packages/python/beagle

~/beagle_venv/bin/mypy ./ 
```

## Python Dependencies

### Required Python Version

Beagle expects to be run with Python version 3.10 or greater.

### Additional Libraries

Some optional CI automation requires `boto3` and `botocore` as specified in the [project requirements](requirements.txt)
file. If these libraries are not present, beagle will output a warning to `stderr` but continue otherwise:

```text
AWS functions disabled: boto3 and botocore python libraries are missing from PYTHONPATH
```

## Beagle Extensions

Beagle supports conditional loading of extensions and will do so when invoked. Extensions are expected to be present in
a top-level repository directory named `beagle/extensions`.

```bash
$ ls beagle/extensions
__init__.py  build
```

### Initialization Hook

Beagle extensions are expected to present a function in the `extensions` namespace named `init`:

```bash
cat example_repo/beagle/extensions/__init__.py
```

```python
from beagle.project import ProjectContext


def init(project_ctx: ProjectContext) -> None:
    # This function will be called upon import of the beagle extensions package. Any additional wire-up may be
    # performed here 
    pass
```

### Setting Test and Build Plans from an Extension

Beagle extensions are responsible for setting up test and build plans prior to execution. Test and build plans are
sourced from the top-level [plans](packages/python/beagle/plans) package presented by beagle:

```bash
PYTHONPATH=packages/python/beagle python

Python 3.10.10 (main, Mar 18 2023, 21:14:51) [GCC 12.2.1 20230121] on linux
Type "help", "copyright", "credits" or "license" for more information.

>>> import plans
>>> dir(plans)
['__builtins__', '__cached__', '__doc__', '__file__', '__loader__', '__name__', '__package__', '__path__', '__spec__', 'build_plans', 'test_plans']

>>> print(plans.test_plans)
[]

>>> print(plans.build_plans)
[]
>>>
```

#### Test and Build Plan Setup Example

The following example is taken from this repository's set of beagle extensions:

```python
# Import global plans so that beagle extensions can add to a registry of prepared plans available to run
import plans

# Supporting beagle imports
from beagle.plan import GolangWorkspaceBuildPlan, GolangWorkspaceTestPlan, CopyBloodHoundUIAssets, YarnTestPlan,

YarnBuildPlan
from beagle.project import ProjectContext


def init(project_ctx: ProjectContext) -> None:
    """
    init is the beagle initializer function. This is the entry point for these particular extensions.
    
    :param project_ctx: the project context contains all context of the beagle environment, runtime and project files
    :return:
    """
    project_ctx.info("Community beagle extensions enabled")

    plans.build_plans = [
        YarnBuildPlan(
            name="bh-ui",
            source_path=project_ctx.fs.project_path("cmd", "ui"),
            project_ctx=project_ctx,
        ),
        GolangWorkspaceBuildPlan(
            name="bh",
            project_ctx=project_ctx,
            prepare_actions=[
                CopyBloodHoundUIAssets(
                    ui_build_name="bh-ui",
                    ce_root_path=project_ctx.fs.project_path(),
                )
            ],
        ),
    ]

    plans.test_plans = [
        YarnTestPlan(
            name="bh-shared-ui",
            source_path=project_ctx.fs.project_path("packages", "javascript", "bh-shared-ui"),
            project_ctx=project_ctx,
        ),
        YarnTestPlan(
            name="js-client-library",
            source_path=project_ctx.fs.project_path("packages", "javascript", "js-client-library"),
            project_ctx=project_ctx,
        ),
        YarnTestPlan(
            name="bh-ui",
            source_path=project_ctx.fs.project_path("cmd", "ui"),
            project_ctx=project_ctx,
        ),
        GolangWorkspaceTestPlan(
            name="bh",
            project_ctx=project_ctx,
        ),
    ]
```
