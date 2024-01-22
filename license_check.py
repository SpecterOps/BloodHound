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

# About this file:
#
# license_check.py is a naive attempt at automatically auditing and inserting copyright headers
# for the chosen license. It will attempt to match the license header and identify if the header
# has been damaged.
#
# Review of changed files is a must before committing after running this file.
#

import os
import pathlib

from typing import List

# Source root files for license information
ROOT_LICENSE_FILE = "LICENSE"
ROOT_LICENSE_HEADER_FILE = "LICENSE.header"

# Apache License 2.0 copy
LICENSE = """                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.

      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.

      "Source" form shall mean the preferred form for making modifications,
      including but not limited to software source code, documentation
      source, and configuration files.

      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form, including but
      not limited to compiled object code, generated documentation,
      and conversions to other media types.

      "Work" shall mean the work of authorship, whether in Source or
      Object form, made available under the License, as indicated by a
      copyright notice that is included in or attached to the work
      (an example is provided in the Appendix below).

      "Derivative Works" shall mean any work, whether in Source or Object
      form, that is based on (or derived from) the Work and for which the
      editorial revisions, annotations, elaborations, or other modifications
      represent, as a whole, an original work of authorship. For the purposes
      of this License, Derivative Works shall not include works that remain
      separable from, or merely link (or bind by name) to the interfaces of,
      the Work and Derivative Works thereof.

      "Contribution" shall mean any work of authorship, including
      the original version of the Work and any modifications or additions
      to that Work or Derivative Works thereof, that is intentionally
      submitted to Licensor for inclusion in the Work by the copyright owner
      or by an individual or Legal Entity authorized to submit on behalf of
      the copyright owner. For the purposes of this definition, "submitted"
      means any form of electronic, verbal, or written communication sent
      to the Licensor or its representatives, including but not limited to
      communication on electronic mailing lists, source code control systems,
      and issue tracking systems that are managed by, or on behalf of, the
      Licensor for the purpose of discussing and improving the Work, but
      excluding communication that is conspicuously marked or otherwise
      designated in writing by the copyright owner as "Not a Contribution."

      "Contributor" shall mean Licensor and any individual or Legal Entity
      on behalf of whom a Contribution has been received by Licensor and
      subsequently incorporated within the Work.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      (except as stated in this section) patent license to make, have made,
      use, offer to sell, sell, import, and otherwise transfer the Work,
      where such license applies only to those patent claims licensable
      by such Contributor that are necessarily infringed by their
      Contribution(s) alone or by combination of their Contribution(s)
      with the Work to which such Contribution(s) was submitted. If You
      institute patent litigation against any entity (including a
      cross-claim or counterclaim in a lawsuit) alleging that the Work
      or a Contribution incorporated within the Work constitutes direct
      or contributory patent infringement, then any patent licenses
      granted to You under this License for that Work shall terminate
      as of the date such litigation is filed.

   4. Redistribution. You may reproduce and distribute copies of the
      Work or Derivative Works thereof in any medium, with or without
      modifications, and in Source or Object form, provided that You
      meet the following conditions:

      (a) You must give any other recipients of the Work or
          Derivative Works a copy of this License; and

      (b) You must cause any modified files to carry prominent notices
          stating that You changed the files; and

      (c) You must retain, in the Source form of any Derivative Works
          that You distribute, all copyright, patent, trademark, and
          attribution notices from the Source form of the Work,
          excluding those notices that do not pertain to any part of
          the Derivative Works; and

      (d) If the Work includes a "NOTICE" text file as part of its
          distribution, then any Derivative Works that You distribute must
          include a readable copy of the attribution notices contained
          within such NOTICE file, excluding those notices that do not
          pertain to any part of the Derivative Works, in at least one
          of the following places: within a NOTICE text file distributed
          as part of the Derivative Works; within the Source form or
          documentation, if provided along with the Derivative Works; or,
          within a display generated by the Derivative Works, if and
          wherever such third-party notices normally appear. The contents
          of the NOTICE file are for informational purposes only and
          do not modify the License. You may add Your own attribution
          notices within Derivative Works that You distribute, alongside
          or as an addendum to the NOTICE text from the Work, provided
          that such additional attribution notices cannot be construed
          as modifying the License.

      You may add Your own copyright statement to Your modifications and
      may provide additional or different license terms and conditions
      for use, reproduction, or distribution of Your modifications, or
      for any such Derivative Works as a whole, provided Your use,
      reproduction, and distribution of the Work otherwise complies with
      the conditions stated in this License.

   5. Submission of Contributions. Unless You explicitly state otherwise,
      any Contribution intentionally submitted for inclusion in the Work
      by You to the Licensor shall be under the terms and conditions of
      this License, without any additional terms or conditions.
      Notwithstanding the above, nothing herein shall supersede or modify
      the terms of any separate license agreement you may have executed
      with Licensor regarding such Contributions.

   6. Trademarks. This License does not grant permission to use the trade
      names, trademarks, service marks, or product names of the Licensor,
      except as required for reasonable and customary use in describing the
      origin of the Work and reproducing the content of the NOTICE file.

   7. Disclaimer of Warranty. Unless required by applicable law or
      agreed to in writing, Licensor provides the Work (and each
      Contributor provides its Contributions) on an "AS IS" BASIS,
      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
      implied, including, without limitation, any warranties or conditions
      of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
      PARTICULAR PURPOSE. You are solely responsible for determining the
      appropriateness of using or redistributing the Work and assume any
      risks associated with Your exercise of permissions under this License.

   8. Limitation of Liability. In no event and under no legal theory,
      whether in tort (including negligence), contract, or otherwise,
      unless required by applicable law (such as deliberate and grossly
      negligent acts) or agreed to in writing, shall any Contributor be
      liable to You for damages, including any direct, indirect, special,
      incidental, or consequential damages of any character arising as a
      result of this License or out of the use or inability to use the
      Work (including but not limited to damages for loss of goodwill,
      work stoppage, computer failure or malfunction, or any and all
      other commercial damages or losses), even if such Contributor
      has been advised of the possibility of such damages.

   9. Accepting Warranty or Additional Liability. While redistributing
      the Work or Derivative Works thereof, You may choose to offer,
      and charge a fee for, acceptance of support, warranty, indemnity,
      or other liability obligations and/or rights consistent with this
      License. However, in accepting such obligations, You may act only
      on Your own behalf and on Your sole responsibility, not on behalf
      of any other Contributor, and only if You agree to indemnify,
      defend, and hold each Contributor harmless for any liability
      incurred by, or claims asserted against, such Contributor by reason
      of your accepting any such warranty or additional liability.

   END OF TERMS AND CONDITIONS

   APPENDIX: How to apply the Apache License to your work.

      To apply the Apache License to your work, attach the following
      boilerplate notice, with the fields enclosed by brackets "[]"
      replaced with your own identifying information. (Don't include
      the brackets!)  The text should be enclosed in the appropriate
      comment syntax for the file format. We also recommend that a
      file or class name and description of purpose be included on the
      same "printed page" as the copyright notice for easier
      identification within third-party archives.

   Copyright [yyyy] [name of copyright owner]

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
"""

# Apache License 2.0 header copy
LICENSE_HEADER = """Copyright 2023 Specter Ops, Inc.

Licensed under the Apache License, Version 2.0
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0"""

# XML requires a block quote, so it's easier to just hand format it here
XML_LICENSE_HEADER = "<!--\n" + LICENSE_HEADER + "\n-->"

# Any path that contains one of these elements will be ignored when performing license checking. Eventually would
# like to have this also extend to the .gitignore file.
IGNORED_PATH_ELEMENTS = [
    "node_modules",
    "dist",
    ".beagle",
    ".yarn",
    "cmd/api/src/api/static/assets",

    # This is generated code that we don't really care about
    "packages/go/cypher/parser",

    # These are vendored packages
    "packages/python/beagle/beagle/semver",
    "cmd/api/src/cmd/testidp/samlidp",

    # Ignore checksums
    "sha256",

    # Ignore this file
    "license_check.py",
]

# Any file extension found below will be ignored when performing license checking.
IGNORED_EXTENSIONS = [
    ".iml",
    ".zip",
    ".g4",
    ".sum",
    ".bazel",
    ".bzl",
    ".typed",
    ".md",
    ".json",
    ".template",
    "sha256",
    ".pyc",
    ".gif",
    ".tiff",
    ".lock",
    ".txt",
    ".png",
    ".jpg",
    ".jpeg",
    ".ico",
    ".gz",
    ".tar",
    ".woff2",
    ".header",
    ".pro",
    ".cert",
    ".crt",
    ".key",
    ".example",
    ".svg",
]

# Any file listed below is included regardless of exclusions.
MUST_CHECK_PATHS = [
    ".yarn/plugins/nested-workspace/plugin-nested-workspace.js"
]


def is_path_ignored(path: str) -> bool:
    return any([ignored_elem in path for ignored_elem in IGNORED_PATH_ELEMENTS])


def generate_license_header(comment_prefix: str) -> str:
    golang_header: str = ""

    for line in LICENSE_HEADER.split("\n"):
        if line.strip() == "":
            golang_header += comment_prefix + "\n"
        else:
            golang_header += comment_prefix + " " + line + "\n"

    return golang_header


# Below is a map of file extension to a correctly commented and formatted license header.
LICENSE_HEADERS_BY_EXTENSION = {
    ".svg": XML_LICENSE_HEADER,
    ".xml": XML_LICENSE_HEADER,
    ".html": XML_LICENSE_HEADER,
    ".go": generate_license_header("//"),
    ".mod": generate_license_header("//"),
    ".work": generate_license_header("//"),
    ".tsx": generate_license_header("//"),
    ".ts": generate_license_header("//"),
    ".js": generate_license_header("//"),
    ".jsx": generate_license_header("//"),
    ".cjs": generate_license_header("//"),
    ".py": generate_license_header("#"),
    ".yaml": generate_license_header("#"),
    ".yml": generate_license_header("#"),
    ".sh": generate_license_header("#"),
    ".scss": generate_license_header("//"),
    ".Dockerfile": generate_license_header("#"),
    ".cue": generate_license_header("//"),
    ".sql": generate_license_header("--"),
    ".toml": generate_license_header("#"),
}

# Below is a list of valid file headers that the license must be placed after
FILE_HEADER_PREFIXES = [
    # POSIX exec header
    "#!",

    # XML header
    "<?xml"
]


def content_has_header(path: str, content_lines: List[str], header: str) -> bool:
    matching_header = False
    header_lines = header.splitlines()
    header_lineno = 0

    for line in content_lines:
        if line.strip() == header_lines[header_lineno].strip():
            # Mark that we're matching the header
            if not matching_header:
                matching_header = True

            # Advance the line being matched
            header_lineno += 1

            # Ignore this file, it has a valid license header
            if header_lineno >= len(header_lines):
                return True

        elif matching_header:
            print(f"WARNING: Path {path} contains damaged license information.")
            return True

    return False


def _is_file_header(line: str) -> bool:
    for header in FILE_HEADER_PREFIXES:
        if line.startswith(header):
            return True
    return False


def insert_license_header(path: str, header: str) -> None:
    with open(path, "r") as fin:
        content = fin.read()

    # Split the content into lines
    content_lines = content.splitlines()

    # Check if there is no content or if the content already has a license header
    if content_has_header(path, content_lines, header):
        return

    # Try to find a script exec header to advance the line offset
    line_offset = 1 if len(content_lines) > 0 and _is_file_header(content_lines[0]) else 0

    for line in content_lines[line_offset:]:
        # Make sure to skip leading newlines since we'll add our own
        if line.strip() != "":
            break

        line_offset += 1

    # Output the file with the copy
    with open(path, "w") as fout:
        if line_offset > 0:
            # Write out the exec header first
            fout.write(content_lines[0])

            # Add a trailing newline to the header to separate it from the license
            fout.write("\n\n")

        # Write out the license header for the file
        fout.write(header)
        fout.write("\n")

        # Finish writing the content line by line starting at the exec header offset, if one exists
        for line in content_lines[line_offset:]:
            fout.write(line)
            fout.write("\n")


def validate_license_header(path: str, header: str):
    with open(path, "r") as fin:
        content = fin.read()

    if content.count("SPDX-License-Identifier: Apache-2.0") != 1:
        print(f"WARNING: License for {path} may be damaged.")


def check_file(path: pathlib.Path, path_str: str) -> None:
    license_header = LICENSE_HEADERS_BY_EXTENSION.get(path.suffix)

    if license_header is None:
        if path.suffix not in IGNORED_EXTENSIONS:
            print(f"No license header information for file extension: {path.suffix} {path}")
    else:
        insert_license_header(path_str, license_header)
        validate_license_header(path_str, license_header)


def main() -> None:
    if not os.path.isfile(ROOT_LICENSE_FILE):
        print(f"{ROOT_LICENSE_FILE} file is missing. Writing a new copy.")

        with open(ROOT_LICENSE_FILE, "w") as fout:
            fout.write(LICENSE)

    if not os.path.isfile(ROOT_LICENSE_HEADER_FILE):
        print(f"{ROOT_LICENSE_HEADER_FILE} file is missing. Writing a new copy.")

        with open(ROOT_LICENSE_HEADER_FILE, "w") as fout:
            fout.write(LICENSE_HEADER)

    # Walk the root first
    for root_path, _, filenames in os.walk("./"):
        for filename in filenames:
            path_str = os.path.join(root_path, filename)
            path = pathlib.Path(path_str)

            if path.suffix is None or path.suffix == "" or is_path_ignored(path_str):
                continue

            check_file(path, path_str)

    # Visit any files that may have been excluded but are marked as must check
    for path_str in MUST_CHECK_PATHS:
        check_file(pathlib.Path(path_str), path_str)


if __name__ == "__main__":
    main()
