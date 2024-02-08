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

FROM docker.io/dpage/pgadmin4

# Add bh server config
COPY configs/pgadmin/servers.json /pgadmin4/servers.json

# Need to make directory
RUN mkdir -p /var/lib/pgadmin/storage/bloodhound_specterops.io/
COPY configs/pgadmin/pgpass /var/lib/pgadmin/storage/bloodhound_specterops.io/pgpass

# Give pgadmin ownership or it will be owned by root and set u(rw) for password file or pgadmin will not use the file
USER root
RUN chown -R pgadmin /var/lib/pgadmin && chmod 600 /var/lib/pgadmin/storage/bloodhound_specterops.io/pgpass
