# Copyright 2025 Specter Ops, Inc.
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

"""
To utilize this example please install requests. The rest of the dependencies are part of the Python 3 standard
library.

# pip install --upgrade requests

Note: this script was written for Python 3.6.X or greater.

Insert your BHE API creds in the BHE constants and change the PRINT constants to print desired data.
"""

import hmac
import hashlib
import base64
import requests
import datetime
import json

from typing import Optional


BHE_DOMAIN = "xyz.bloodhoundenterprise.io"
BHE_PORT = 443
BHE_SCHEME = "https"
BHE_TOKEN_ID = "xxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
BHE_TOKEN_KEY = ""

PRINT_PRINCIPALS = False
PRINT_ATTACK_PATH_TIMELINE_DATA = False
PRINT_POSTURE_DATA = False

DATA_START = "1970-01-01T00:00:00.000Z"
DATA_END = datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%S.%f')[:-3] + 'Z' # Now

class Credentials(object):
    def __init__(self, token_id: str, token_key: str) -> None:
        self.token_id = token_id
        self.token_key = token_key


class APIVersion(object):
    def __init__(self, api_version: str, server_version: str) -> None:
        self.api_version = api_version
        self.server_version = server_version


class Domain(object):
    def __init__(self, name: str, id: str, collected: bool, domain_type: str, impact_value: int) -> None:
        self.name = name
        self.id = id
        self.type = domain_type
        self.collected = collected
        self.impact_value = impact_value


class AttackPath(object):
    def __init__(self, id: str, title: str, domain: Domain) -> None:
        self.id = id
        self.title = title
        self.domain_id = domain.id
        self.domain_name = domain.name.strip()

    def __lt__(self, other):
        return self.exposure < other.exposure


class Client(object):
    def __init__(self, scheme: str, host: str, port: int, credentials: Credentials) -> None:
        self._scheme = scheme
        self._host = host
        self._port = port
        self._credentials = credentials

    def _format_url(self, uri: str) -> str:
        formatted_uri = uri
        if uri.startswith("/"):
            formatted_uri = formatted_uri[1:]

        return f"{self._scheme}://{self._host}:{self._port}/{formatted_uri}"

    def _request(self, method: str, uri: str, body: Optional[bytes] = None) -> requests.Response:
        # Digester is initialized with HMAC-SHA-256 using the token key as the HMAC digest key.
        digester = hmac.new(self._credentials.token_key.encode(), None, hashlib.sha256)

        # OperationKey is the first HMAC digest link in the signature chain. This prevents replay attacks that seek to
        # modify the request method or URI. It is composed of concatenating the request method and the request URI with
        # no delimiter and computing the HMAC digest using the token key as the digest secret.
        #
        # Example: GET /api/v1/test/resource HTTP/1.1
        # Signature Component: GET/api/v1/test/resource
        digester.update(f"{method}{uri}".encode())

        # Update the digester for further chaining
        digester = hmac.new(digester.digest(), None, hashlib.sha256)

        # DateKey is the next HMAC digest link in the signature chain. This encodes the RFC3339 formatted datetime
        # value as part of the signature to the hour to prevent replay attacks that are older than max two hours. This
        # value is added to the signature chain by cutting off all values from the RFC3339 formatted datetime from the
        # hours value forward:
        #
        # Example: 2020-12-01T23:59:60Z
        # Signature Component: 2020-12-01T23
        datetime_formatted = datetime.datetime.now().astimezone().isoformat("T")
        digester.update(datetime_formatted[:13].encode())

        # Update the digester for further chaining
        digester = hmac.new(digester.digest(), None, hashlib.sha256)

        # Body signing is the last HMAC digest link in the signature chain. This encodes the request body as part of
        # the signature to prevent replay attacks that seek to modify the payload of a signed request. In the case
        # where there is no body content the HMAC digest is computed anyway, simply with no values written to the
        # digester.
        if body is not None:
            digester.update(body)

        # Perform the request with the signed and expected headers
        return requests.request(
            method=method,
            url=self._format_url(uri),
            headers={
                "User-Agent": "bhe-python-sdk 0001",
                "Authorization": f"bhesignature {self._credentials.token_id}",
                "RequestDate": datetime_formatted,
                "Signature": base64.b64encode(digester.digest()),
                "Content-Type": "application/json",
            },
            data=body,
        )

    def get_version(self) -> APIVersion:
        response = self._request("GET", "/api/version")
        payload = response.json()

        return APIVersion(api_version=payload["data"]["API"]["current_version"], server_version=payload["data"]["server_version"])

    def get_domains(self) -> list[Domain]:
        response = self._request('GET', '/api/v2/available-domains')
        payload = response.json()['data']

        domains = list()
        for domain in payload:
            domains.append(Domain(domain["name"], domain["id"], domain["collected"], domain["type"], domain["impactValue"]))

        return domains

    def get_paths(self, domain: Domain) -> list:
        response = self._request('GET', '/api/v2/domains/' + domain.id + '/available-types')
        path_ids = response.json()['data']

        paths = list()
        for path_id in path_ids:
            # Get nice title from API and strip newline
            path_title = self._request('GET', '/ui/findings/' + path_id + '/title.md')

            # Create attackpath object
            path = AttackPath(path_id, path_title.text.strip(), domain)
            paths.append(path)

        return paths

    def get_path_principals(self, path: AttackPath) -> list:
        # Get path details from API
        response = self._request('GET', '/api/v2/domains/' + path.domain_id + '/details?finding=' + path.id + '&skip=0&limit=0&Accepted=eq:False')
        payload = response.json()

        # Build dictionary of impacted pricipals
        if 'count' in payload:
            path.impacted_principals = list()
            for path_data in payload['data']:
                # Check for both From and To to determine whether relational or configuration path
                if (path.id.startswith('LargeDefault')):
                    from_principal = path_data['FromPrincipalProps']['name']
                    to_principal = path_data['ToPrincipalProps']['name']
                    principals = {
                        'Group': from_principal,
                        'Principal': to_principal
                    }
                elif ('FromPrincipalProps' in path_data) and ('ToPrincipalProps' in path_data):
                    from_principal = path_data['FromPrincipalProps']['name']
                    to_principal = path_data['ToPrincipalProps']['name']
                    principals = {
                        'Non Tier Zero Principal': from_principal,
                        'Tier Zero Principal': to_principal
                    }
                else:
                    principals = {
                        'User': path_data['Props']['name']
                    }
                path.impacted_principals.append(principals)
                path.principal_count = payload['count']
        else:
            path.principal_count = 0

        return path

    def get_path_timeline(self, path: AttackPath, from_timestamp: str, to_timestamp: str):
        # Sparkline data
        response = self._request('GET', '/api/v2/domains/' + path.domain_id + '/sparkline?finding=' + path.id + '&from=' + from_timestamp + '&to=' + to_timestamp)
        exposure_data = response.json()['data']

        events = list()
        for event in exposure_data:
            e = {}
            e['finding_id'] = path.id
            e['domain_id'] = path.domain_id
            e['path_title'] = path.title
            e['exposure'] = event['CompositeRisk']
            e['finding_count'] = event['FindingCount']
            e['principal_count'] = event['ImpactedAssetCount']
            e['id'] = event['id']
            e['created_at'] = event['created_at']
            e['updated_at'] = event['updated_at']
            e['deleted_at'] = event['deleted_at']

            # Determine severity from exposure
            e['severity'] = self.get_severity(e['exposure'])
            events.append(e)

        return events

    def get_posture(self, from_timestamp: str, to_timestamp: str) -> list:
        response = self._request('GET', '/api/v2/posture-stats?from=' + from_timestamp + '&to=' + to_timestamp)
        payload = response.json()
        return payload["data"]

    def get_severity(self, exposure: int) -> str:
        severity = 'Low'
        if exposure > 40: severity = 'Moderate'
        if exposure > 80: severity = 'High'
        if exposure > 95: severity = 'Critical'
        return severity
    
    def run_cypher(self, query, include_properties=False) -> requests.Response:
        """ Runs a Cypher query and returns the results

        Parameters:
        query (string): The Cypher query to run
        include_properties (bool): Should all properties of result nodes/edges be returned

        Returns:
        string: JSON result

        """

        data = {
            "include_properties": include_properties,
            "query": query
        }
        body = json.dumps(data).encode('utf8')
        response = self._request("POST", "/api/v2/graphs/cypher", body)
        return response.json()

def main() -> None:
    # This might be best loaded from a file
    credentials = Credentials(
        token_id=BHE_TOKEN_ID,
        token_key=BHE_TOKEN_KEY,
    )

    # Create the client and perform an example call using token request signing
    client = Client(scheme=BHE_SCHEME, host=BHE_DOMAIN, port=BHE_PORT, credentials=credentials)
    version = client.get_version()

    print("BHE Python API Client Example")
    print(f"API version: {version.api_version} - Server version: {version.server_version}\n")

    domains = client.get_domains()

    print("Available Domains")
    for domain in domains:
        print(f"* {domain.name} (id: {domain.id}, collected: {domain.collected}, type: {domain.type}, exposure: {domain.impact_value})")
    
    # Cypher query for Kerberoastable users
    print("Kerberoastable Users")
    cypher = "MATCH (n:User) WHERE n.hasspn=true RETURN n"
    cypher_result = client.run_cypher(cypher)
    # Get nodes from Cypher result
    nodes = cypher_result['data']['nodes']
    if cypher_result['data']['nodes']:
        for node_id, node_data in nodes.items():
            print(node_data['label'])

    # Display paths in each domain
    for domain in domains:
        if domain.collected:
            # Get paths for domain
            attack_paths = client.get_paths(domain)
            print(("\nProcessing %s attack paths for domain %s" % (len(attack_paths), domain.name)))

            for attack_path in attack_paths:
                print("Processing attack path %s" % attack_path.id)

                # Get attack path principals
                if (PRINT_PRINCIPALS):
                    path_principals = client.get_path_principals(attack_path)
                    print(path_principals.__dict__)

                # Get attack path timeline
                if (PRINT_ATTACK_PATH_TIMELINE_DATA):
                    path_events = client.get_path_timeline(
                        path = attack_path,
                        from_timestamp = DATA_START,
                        to_timestamp = DATA_END
                    )
                    print(path_events)

    # Get posture data
    if (PRINT_POSTURE_DATA):
        posture_events = client.get_posture(
            from_timestamp = DATA_START,
            to_timestamp = DATA_END
        )
        print("%s events of posture data" % len(posture_events))
        print(posture_events)


if __name__ == "__main__":
    main()
