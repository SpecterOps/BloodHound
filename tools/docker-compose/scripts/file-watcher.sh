#!/usr/bin/env bash

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

while getopts "hsf:c:" flag
do
  case $flag in
    h)
      echo "usage: file-watcher.sh [-h][-s][-f <awk filter string>][-c <command>]
      
      -h this help message
      -s set git safe.directory
      -c command to run with a file-watcher
      -f awk filter to use for the file list
      "
      ;;
    f)
      FILTER=$OPTARG
      ;;
    c)
      COMMAND=$OPTARG
      ;;
  esac
done

while sleep 0.5 # gives time for user to interrupt twice
do
  echo Running $COMMAND with filter $FILTER
  # Recurse submodules works for cached, but untracked has to be done separately
  # The submodule foreach gets all the untracked/nonignored files, applies our relative path to them, and appends them to the list of files to watch
  (
    git ls-files --recurse-submodules;
    git ls-files . --exclude-standard --others;
    git submodule --quiet foreach 'export path;bash -c '\''git ls-files --others --exclude-standard | sed "s/^/${path/\//\\/}\//"'\'
  ) | awk "$FILTER" | entr -drn $COMMAND
done
