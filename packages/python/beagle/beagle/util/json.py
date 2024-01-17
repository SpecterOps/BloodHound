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
import json

from typing import List, Any, TextIO, Generator, Union, Dict


def json_multi_loads(reader: TextIO) -> Generator[Union[List[Any], Dict[str, Any]], None, None]:
    """
    json_multi_loads is a quasi-parser that takes a reader which emits one or multiple JSON arrays or objects and yields
    a generator of the parsed JSON objects. Upon failure to parse a top-level JSON element this function raaises an
    the resulting exception.

    :param reader: reader to read raw JSON data from
    :return: all parsed JSON elements
    :raises: json.JSONDecodeError
    """

    buffer = io.StringIO()
    in_string = False
    escaping = False
    depth = 0

    for line in reader.readlines():
        for next_char in line:
            if next_char == "\\" and in_string:
                escaping = True
                continue

            buffer.write(next_char)

            if escaping:
                escaping = False

            elif next_char == '"':
                in_string = not in_string

            elif next_char in ("{", "["):
                depth += 1

            elif next_char in ("}", "]"):
                depth -= 1

                if depth == 0:
                    buffer.seek(0)
                    yield json.loads(buffer.read())

                    buffer.seek(0)
                    buffer.truncate()
