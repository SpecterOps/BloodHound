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

FROM docker.io/library/neo4j:4.4.0 as neo4j
RUN echo "dbms.security.auth_enabled=false" >> /var/lib/neo4j/conf/neo4j.conf
RUN neo4j-admin memrec >> /var/lib/neo4j/conf/neo4j.conf
