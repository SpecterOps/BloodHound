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

FROM docker.io/library/neo4j:4.4.32 as base

ARG memconfig

RUN echo "dbms.security.auth_enabled=false" >> /var/lib/neo4j/conf/neo4j.conf && \
    # Restrict allowed procedures only to what is used to mitigate CVE-2023-23926
    echo "dbms.security.procedures.unrestricted=apoc.periodic.*,specterops.*" >> /var/lib/neo4j/conf/neo4j.conf && \
    echo "dbms.security.procedures.allowlist=apoc.periodic.*,specterops.*" >> /var/lib/neo4j/conf/neo4j.conf

RUN if [ "$memconfig" = "true" ]; then neo4j-admin memrec >> /var/lib/neo4j/conf/neo4j.conf; fi

RUN cp /var/lib/neo4j/labs/apoc-4.4.0.26-core.jar /var/lib/neo4j/plugins/apoc-4.4.0.26-core.jar
