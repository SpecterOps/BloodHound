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

import io
import os
import subprocess
import sys

from threading import Thread
from typing import List, TextIO, IO, Dict, Any, Optional


class CommandException(Exception):
    def __init__(self, formatted_command: str, returncode: int, output: TextIO, cwd: str) -> None:
        self.formatted_command = formatted_command
        self.returncode = returncode
        self.output = output
        self.cwd = cwd


# Read output until the program finishes
def _tee(reader: IO[bytes], writers: List[TextIO]):
    for line_bytes in reader:
        decoded_line = line_bytes.decode("utf-8")

        for writer in writers:
            writer.write(decoded_line)


def run_simple(cmd: List[str], cwd: str, environment: Optional[Dict[str, Any]] = None) -> int:
    # Default to passing through os.environ if the environment isn't set
    subprocess_env = dict(**os.environ)

    if environment is not None:
        subprocess_env = environment

    # Dump all output and return the subprocess return code only
    process = subprocess.Popen(
        args=cmd,
        cwd=cwd,
        env=subprocess_env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    process.wait()
    return process.returncode


def _run(
    cmd: List[str],
    cwd: str,
    capture_stderr: bool = False,
    environment: Optional[Dict[str, Any]] = None,
    writers: List[TextIO] = list(),
) -> int:
    # Default to passing through os.environ if the environment isn't set
    subprocess_env = dict(**os.environ)

    if environment is not None:
        subprocess_env = environment

    # Use a pipe and redirect stderr to stdout so that we can fork output both to a file and to the console
    process = subprocess.Popen(
        args=cmd,
        cwd=cwd,
        env=subprocess_env,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT if capture_stderr else subprocess.DEVNULL,
    )

    # Spin up an output thread to make sure the process doesn't block on reading
    output_thread = Thread(target=_tee, args=[process.stdout, writers])
    output_thread.start()

    # Wait for the process to exit and then join on the output thread
    process.wait()
    output_thread.join()

    return process.returncode


def run(
    cmd: List[str],
    cwd: str,
    capture_stderr: bool = False,
    environment: Optional[Dict[str, Any]] = None,
    additional_writers: List[TextIO] = list(),
) -> TextIO:
    capture_writer = io.StringIO()
    returncode = _run(cmd, cwd, capture_stderr, environment, [capture_writer] + additional_writers)

    # Rewind the StringIO buffer
    capture_writer.seek(0)

    # Validate the return code to ensure that the subprocess succeeded
    if returncode != 0:
        raise CommandException(
            formatted_command=" ".join(cmd),
            returncode=returncode,
            output=capture_writer,
            cwd=cwd,
        )

    return capture_writer


def run_logged(
    cmd: List[str], cwd: str, log_path: str, print_output: bool = False, environment: Optional[Dict[str, Any]] = None
) -> TextIO:
    with open(log_path, "a") as log_writer:
        if not print_output:
            return run(cmd=cmd, cwd=cwd, capture_stderr=True, environment=environment, additional_writers=[log_writer])
        else:
            return run(
                cmd=cmd,
                cwd=cwd,
                capture_stderr=True,
                environment=environment,
                additional_writers=[log_writer, sys.stdout],
            )
