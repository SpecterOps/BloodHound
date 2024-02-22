# Copyright 2024 Specter Ops, Inc.
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

import json
import sys

with open(sys.argv[1], 'r') as f:
    j = json.load(f)

nodes = {}
relationships = []


class BaseNode:
    def __init__(self, name: str):
        self.name = name


class UserNode(BaseNode):
    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryUser("{self.name}", domainSid)'


class GroupNode(BaseNode):
    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryGroup("{self.name}", domainSid)'


class ComputerNode(BaseNode):
    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryComputer("{self.name}", domainSid)'


class OUNode(BaseNode):
    def __init__(self, name: str, blocksInheritance: bool = False):
        super().__init__(name)
        self.blocksInheritance = blocksInheritance

    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryOU("{self.name}", domainSid, {json.dumps(self.blocksInheritance)})'


class DomainNode(BaseNode):
    def __init__(self, name: str, blocksInheritance: bool = False):
        super().__init__(name)
        self.blocksInheritance = blocksInheritance

    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryDomain("{self.name}", domainSid, {json.dumps(self.blocksInheritance)}, true)'


class NTAuthStore(BaseNode):
    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryNTAuthStore("{self.name}", domainSid)'


class RootCA(BaseNode):
    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryRootCA("{self.name}", domainSid)'


class EnterpriseCA(BaseNode):
    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.NewActiveDirectoryEnterpriseCA("{self.name}", domainSid)'


class CertTemplateData:
    def __init__(self) -> None:
        self.RequiresManagerApproval = False
        self.AuthenticationEnabled = False
        self.EnrolleeSuppliesSubject = False
        self.SubjectAltRequireUPN = False
        self.SubjectAltRequireSPN = False
        self.NoSecurityExtension = False
        self.SchemaVersion = 1
        self.AuthorizedSignatures = 0
        self.EKUS = []
        self.ApplicationPolicies = []
        self.SubjectAltRequireEmail = False

    def toJSON(self):
        return json.dumps(self, default=lambda o: o.__dict__,
                          sort_keys=True, indent=4)


def create_certtemplate_data(j):
    data = CertTemplateData()
    properties = j['properties']
    for key in properties:
        v = properties[key]
        k = key
        if k == 'RequireManagerApproval':
            k = 'RequiresManagerApproval'
        if v == 'True':
            data.__dict__[k] = True
        elif v == 'False':
            data.__dict__[k] = False
        else:
            try:
                i = int(v)
                data.__dict__[k] = i
            except:
                data.__dict__[k] = v

    return data


class CertTemplate(BaseNode):
    def __init__(self, name: str, data: CertTemplateData):
        super().__init__(name)
        self.data = data

    def create_creation_statement(self):
        d = self.data.toJSON()
        d = d.replace(']', '}')
        d = d.replace('[', '[]string{')
        d = d.replace('"', '')
        d = d[:-2] + "," + d[-2:]
        b = f's.{self.name} = graphTestContext.NewActiveDirectoryCertTemplate("{self.name}", domainSid, CertTemplateData{d})'

        return b


class UnknownNode(BaseNode):
    def create_creation_statement(self):
        return f's.{self.name} = graphTestContext.REPLACEME("{self.name}", domainSid)'


class Relationship:
    def __init__(self, fromID: str, toID: str, type: str) -> None:
        self.fromID = fromID
        self.toID = toID
        self.type = type

    def create_statement(self):
        return f'graphTestContext.NewRelationship(s.{nodes[self.fromID].name}, s.{nodes[self.toID].name}, ad.{self.type})'


for node in j['nodes']:
    name = node['caption']
    id = node['id']
    if 'User' in name:
        nodes[id] = UserNode(name)
    elif 'Group' in name:
        nodes[id] = GroupNode(name)
    elif 'Computer' or 'DC' in name:
        nodes[id] = ComputerNode(name)
    elif 'OU' in name:
        nodes[id] = OUNode(name)
    elif 'NTAuthStore' in name:
        nodes[id] = NTAuthStore(name)
    elif 'RootCA' in name:
        nodes[id] = RootCA(name)
    elif 'Domain' in name:
        nodes[id] = DomainNode(name)
    elif 'EnterpriseCA' in name or 'ECA' in name:
        nodes[id] = EnterpriseCA(name)
    elif 'CertTemplate' in name:
        d = create_certtemplate_data(node)
        nodes[id] = CertTemplate(name, d)
    else:
        print(f'Could not determine type for {name}')
        nodes[id] = UnknownNode(name)

for rel in j['relationships']:
    relationships.append(Relationship(rel['fromId'], rel['toId'], rel['type']))

harnessName = input("Name this harness: ")
structDef = f"type {harnessName} struct {{\n"
setupFunc = f"func (s *{harnessName}) Setup(graphTestContext *GraphTestContext) {{\n domainSid := RandomDomainSID()\n"

for k in {k: v for k, v in sorted(nodes.items(), key=lambda item: item[1].name)}:
    structDef += f'{nodes[k].name} *graph.Node\n'
    setupFunc += nodes[k].create_creation_statement() + "\n"

structDef += "}"

for r in relationships:
    setupFunc += r.create_statement() + "\n"

setupFunc += "}"

print()
print(structDef)
print()
print(setupFunc)
