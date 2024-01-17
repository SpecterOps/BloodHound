#!/usr/bin/env python

# Copyright 2023 Specter Ops, Inc.
# 
# Licensed under the Apache License, Version 2.0
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# 
# SPDX-License-Identifier: Apache-2.0

import argparse
import sys

from beagle import run_builds, run_tests, show_plans
from beagle.project import ProjectContext, new_project_context
from beagle.plan import TestException
from beagle.subprocess import CommandException


def parse_args():
    arg_parser = argparse.ArgumentParser(
        prog="beagle",
        description="Beagle: the BloodHound build, test and packaging automation framework.",
    )

    arg_parser.add_argument(
        "action", default="help", help="action for beagle to pursue", choices=["test", "build", "show"]
    )

    arg_parser.add_argument(
        "-c",
        "--ci",
        default=False,
        action="store_true",
        help="informs beagle that it is running in a CI container",
    )

    arg_parser.add_argument(
        "-d",
        "--disable-generation",
        default=False,
        action="store_true",
        help="informs beagle to not run code generation steps",
    )

    arg_parser.add_argument(
        "-u",
        "--upload-coverage",
        default=False,
        action="store_true",
        help="uploads a new copy of coverage profiles to S3. This flag is only used for CI runs.",
    )

    arg_parser.add_argument(
        "-f",
        "--foss-only",
        default=False,
        action="store_true",
        help="forces the build context to scope down to the BloodHound Community repository only.",
    )

    arg_parser.add_argument(
        "-a",
        "--all-tests",
        default=False,
        action="store_true",
        help="run all tests regardless of staged changes",
    )

    arg_parser.add_argument(
        "-v",
        "--verbose",
        default=False,
        action="store_true",
        help="enable extra output",
    )

    arg_parser.add_argument(
        "-i",
        "--integration",
        default=False,
        action="store_true",
        help="enable integration tests",
    )

    return arg_parser.parse_known_args()


def _run_action(args: argparse.Namespace, project_ctx: ProjectContext) -> None:
    if args.action == "test":
        run_tests(project_ctx=project_ctx, targeted=not args.all_tests)

    elif args.action == "build":
        run_builds(project_ctx=project_ctx)

    elif args.action == "show":
        show_plans(project_ctx=project_ctx)


def main() -> None:
    # Parse args
    args, targets = parse_args()

    # Create the workspace context
    project_ctx = new_project_context(
        local_build=not args.ci,
        verbose=args.verbose,
        run_integration_tests=args.integration,
        upload_coverage=args.upload_coverage,
        targets=targets,
        foss_only=args.foss_only,
        do_code_generation=not args.disable_generation,
    )

    try:
        # Load beagle extensions
        project_ctx.load_beagle_extensions()

        # Ensure the project context is correct
        project_ctx.setup()

        # Run the requested actions
        _run_action(args=args, project_ctx=project_ctx)
    except KeyboardInterrupt:
        # Ignore output for keyboard interrupts - this means a user or some other actor sent beagle a SIGINT signal
        pass
    except TestException as te:
        print(f"Test failed: {te.msg}")

        # Test exceptions are raised as retcode 2
        sys.exit(2)
    except CommandException as ce:
        if not project_ctx.runtime.verbose:
            print(
                f"Command '{ce.formatted_command}' in directory {ce.cwd} returned a non-zero exit code: {ce.returncode}"
            )
        else:
            print(
                f"{ce.output.read().strip()}\nCommand '{ce.formatted_command}' in directory {ce.cwd} returned a non-zero exit code: {ce.returncode}"
            )

        # Generic command exceptions are raised as retcode 1
        sys.exit(1)
    finally:
        project_ctx.teardown()


if __name__ == "__main__":
    main()
