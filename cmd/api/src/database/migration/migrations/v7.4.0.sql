-- Copyright 2025 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE asset_group_history
	ADD COLUMN IF NOT EXISTS email VARCHAR(330) DEFAULT NULL;

-- Populate email for existing records by looking up the email address from the users table
UPDATE asset_group_history
	SET email = (SELECT email_address FROM users WHERE asset_group_history.actor = users.id)
	WHERE email IS NULL AND actor != 'SYSTEM';

-- Add asset_group_tag_selector_nodes table
CREATE TABLE IF NOT EXISTS asset_group_tag_selector_nodes
(
	selector_id int NOT NULL,
	node_id bigint NOT NULL,
	certified int NOT NULL DEFAULT 0,
	certified_by text,
	source int,
	created_at timestamp with time zone,
	updated_at timestamp with time zone,
	CONSTRAINT fk_asset_group_tag_selectors_asset_group_tag_selector_nodes FOREIGN KEY (selector_id) REFERENCES asset_group_tag_selectors(id) ON DELETE CASCADE,
	PRIMARY KEY (selector_id, node_id)
);

-- Add custom_node_kinds table
CREATE TABLE IF NOT EXISTS custom_node_kinds (
  id            SERIAL        PRIMARY KEY,
  kind_name     VARCHAR(256)  NOT NULL,
  config        JSONB         NOT NULL,

  created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

  unique(kind_name)
);

-- Migrate existing Tier Zero selectors
WITH inserted_selector AS (
  INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
  SELECT (SELECT id FROM asset_group_tags WHERE name = 'Tier Zero'), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, false
  FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
  WHERE ag.tag = 'admin_tier_0' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

-- Migrate existing Owned selectors
WITH inserted_kind AS (
  INSERT INTO kind (name)
  SELECT 'Tag_' || replace(name, ' ', '_') as name
  FROM asset_groups
  WHERE tag = 'owned'
  ON CONFLICT DO NOTHING
  RETURNING id, name
),
inserted_tag AS (
  INSERT INTO asset_group_tags (kind_id, type, name, description, created_at, created_by, updated_at, updated_by)
  SELECT ik.id, 3, ag.name, ag.name, current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM'
  FROM inserted_kind ik JOIN asset_groups ag ON ik.name = 'Tag_' || replace(ag.name, ' ', '_')
  RETURNING id, name
),
inserted_selector AS (
  INSERT INTO asset_group_tag_selectors (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
  SELECT (SELECT id from inserted_tag), current_timestamp, 'SYSTEM', current_timestamp, 'SYSTEM', s.name, s.selector, false, true, false
  FROM asset_group_selectors s JOIN asset_groups ag ON ag.id = s.asset_group_id
  WHERE ag.tag = 'owned' and NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = s.name)
  RETURNING id, description
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value) SELECT id, 1, description FROM inserted_selector;

-- Populate default cypher selectors
-- Add the following to the GA release migration to enable these for bootstrapped instances
-- UPDATE asset_group_tag_selectors SET disabled_at = NULL, disabled_by = NULL WHERE is_default = true AND created_at > current_timestamp - '1 min'::interval

WITH src_data AS (
	SELECT * FROM (VALUES
-- START
('Application Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''9B895D92-2CD3-44C7-9D02-A6AC2D5EA5C3@''\nRETURN n;', E'The Application Administrator role can control tenant-resident apps. This includes creating new credentials for apps, which can be used to authenticate the tenant as the app''s service principal and abuse the service principal privileges. The role is therefore considered Tier Zero if the tenant contains any Tier Zero service principals.'),
('Knowledge Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''B5A8DCF3-09D5-43A9-A639-8E29EF291470@''\nRETURN n;', E'The Knowledge Administrator role can control non-role-assignable groups. If any non-role-assignable group has compromising permissions over a Tier Zero asset (e.g. Contributor on a domain controller Azure VM), then the Knowledge Administrator role can add arbitrary principals to the given group and compromise Tier Zero. If no non-role-assignable group has compromising permissions over a Tier Zero asset, then there is no attack path to Tier Zero from the Knowledge Administrator role. It therefore depends on the usage of non-role-assignable groups whether the role should be considered Tier Zero.'),
('Account Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-548''\nRETURN n;', E'The Account Operators group has GenericAll in the default security descriptor on the AD object classes: User, Group, and Computer. That means all objects of these types will be under full control of Account Operators unless they are protected with AdminSDHolder. Not all Tier Zero objects will be protected with AdminSDHolder typically, as not all Tier Zero objects will be included in Protected Accounts and Groups. This means Account Operators members have a path to compromise Tier Zero most often.\n\nIt is possible to delete all GenericAll ACEs for Account Operators on Tier Zero objects. To protect future Tier Zero objects, one would have to either remove the Account Operators ACE from the default security descriptors or implement a process of removing the ACEs as Tier Zero objects are being created. However, we recommend not using the group and classifying it as Tier Zero instead.'),
('Administrators', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-544''\nRETURN n;', E'The Administrators group has full control over most of AD''s essential objects and are inarguably part of Tier Zero.'),
('Backup Operators', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-551''\nRETURN n;', E'The Backup Operators group has the SeBackupPrivilege and SeRestorePrivilege rights on the domain controllers by default. These privileges allow members to access all files on the domain controllers, regardless of their permission, through backup and restore operations. Additionally, Backup Operators have full remote access to the registry of domain controllers. To compromise the domain, members of Backup Operators can dump the registry hives of a domain controller remotely, extract the domain controller account credentials, and perform a DCSync attack. Alternative ways to compromise the domain exist as well. The group is considered Tier Zero because of these known abuse techniques.'),
('Domain Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-512''\nRETURN n;', E'The Domain Admins group has full control over most of AD''s essential objects and are inarguably part of Tier Zero.'),
('Enterprise Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-519''\nRETURN n;', E'The Enterprise Admins group has full control over most of AD''s essential objects and are inarguably part of Tier Zero.'),
('Print Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-550''\nRETURN n;', E'The Print Operators group has the local privilege on the domain controllers to load device drivers and can log on locally on domain controllers by default.\n\nIt is feasible to remove the logon privilege from the group on the domain controllers, such that the group has no known abusable path to Tier Zero. However, the local privilege to load device drivers is considered a security dependency for the domain controllers, and the group is therefore considered Tier Zero.'),
('Schema Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-518''\nRETURN n;', E'The Schema Admins group has full control over the AD schema. This allows the group members to create or modify ACEs for future AD objects. An attacker could grant full control to a compromised principal on any object type and wait for the next Tier Zero asset to be created, to then have a path to Tier Zero. This attack could be remediated by removing any unwanted ACEs on objects before they are promoted to Tier Zero, but we recommend considering the group as Tier Zero instead.'),
('Server Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-549''\nRETURN n;', E'The Server Operators group has local privileges on the domain controllers and perform administrative operations as creating backups of all files. The group can log on locally on domain controllers by default.\n\nIt is feasible to remove the logon privilege from the group on the domain controllers, such that the group has no known abusable path to Tier Zero. However, the local privileges are considered security dependencies for the domain controllers, and the groups are therefore considered Tier Zero.'),
('Administrator', true, false, E'MATCH (n:User)\nWHERE n.objectid ENDS WITH ''-500''\nRETURN n;', E'The built-in Administrator account has admin access to DCs by default and is therefore Tier Zero.'),
('AdminSDHolder', true, false, E'MATCH (n:Domain)\nMATCH (m:Container)\nWHERE m.distinguishedname = ''CN=ADMINSDHOLDER,CN=SYSTEM,'' + n.distinguishedname\nRETURN m;', E'The permissions configured on AdminSDHolder is a template that will be applied on Protected Groups and Users with SDProp, by default every hour. Control over AdminSDHolder means you have control over the Protected Groups (and their members) and Users, which include Tier Zero groups such as Domain Admins. The AdminSDHolder container is therefore a Tier Zero object.'),
('Domain root object', true, false, E'MATCH (n:Domain)\nRETURN n;', E'An attacker with control over the domain root object can compromise the domain in multiple ways, for example by a DCSync attack (see reference). The domain root object is therefore Tier Zero.'),
('KRBTGT objects', false, true, E'MATCH (n:User)\nWHERE n.objectid ENDS WITH ''-502''\nRETURN n;', E'The krbtgt''s credentials allow one to create golden ticket and compromise the domain. Therefore, if you obtain the credentials of this account, then you can authenticate as any Tier Zero user. However, there is currently no known privilege on the object to obtain the Kerberos keys or to compromise the account in any other way. When you reset the password of krbtgt, AD will ignore your password input and use a random string instead. So, the reset password privilege does not work for a compromise. An attacker could use the reset password privilege to harm Tier Zero, as a double password reset causes all Kerberos TGTs in the domain to become invalid. So, since control over the account can harm Tier Zero, and there is no reason for delegating control to non-Tier Zero, the krbtgt is Tier Zero.'),
('Read-Only Domain Controllers', false, true, E'MATCH (n:Computer)-[:MemberOf]->(m:Group) \nWHERE m.objectid ENDS WITH ''-521''\nRETURN n;', E'An attacker with control over a RODC computer object can compromise Tier Zero principals. The attacker can modify the msDS-RevealOnDemandGroup and msDS-NeverRevealGroup attributes of the RODC computer object such that the RODC can retrieve the credentials of a targeted Tier Zero principal. The attacker can obtain admin access to the OS of the RODC through the managedBy attribute, from where they can obtain the credentials of the RODC krbtgt account. With that, the attacker can create a RODC golden ticket for the target principal. This ticket can be converted to a real golden ticket as the target has been added to the msDS-RevealOnDemandGroup attribute and is not protected by the msDS-NeverRevealGroup attribute. Therefore, the RODC computer object is Tier Zero.'),
('Global Administrator', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''62E90394-69F5-4237-9190-012177145E10@''\nRETURN n;', E'The Global Administrator role is the highest privilege role in Entra ID and inarguably part of Tier Zero. It can do almost anything, and grant permission to do the things it cannot do.'),
('Partner Tier2 Support', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''E00E864A-17C5-4A4B-9C06-F5B95A8D5BD8@''\nRETURN n;', E'The Partner Tier2 Support role can reset the password for any principal, including principals with the Global Administrator role. The role is therefore considered Tier Zero.'),
('Privileged Authentication Administrator', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''7BE44C8A-ADAF-4E2A-84D6-AB2649E08A13@''\nRETURN n;', E'The Privileged Authentication Administrator role can set or reset any authentication method (including passwords) for any principal, including principals with the Global Administrator role. The role is therefore considered Tier Zero.'),
('Privileged Role Administrator', true, false, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''E8611AB8-C189-46E8-94E1-60213AB1F814@''\nRETURN n;', E'The Privileged Role Administrator role can grant any other admin role to any principal at the tenant level. The role is therefore considered Tier Zero.'),
('Enterprise CA Computers', false, true, E'MATCH (n:Computer)-[:HostsCAService]->(:EnterpriseCA)\nRETURN n;', E'Enterprise CAs can by default issue certificates that enable authentication as anyone, thereby allowing takeover of Tier Zero. An attacker with admin rights on an enterprise CA can obtain a certificate as any user in different ways. One option is to dump the private key of the CA and craft a ''golden certificate'' as a target user. This attack can be prevented by protecting the private key with hardware. Alternatively, the attacker can publish any template, modify pending certificate requests, and issue denied requests, which typically also enable a takeover of Tier Zero. Enterprise CA computer objects are therefore Tier Zero.\n\nIf the enterprise CA certificate is removed from the NTAuth store, then certificates from this CA cannot be used for domain authentication, thus preventing a Tier Zero takeover.'),
('Exchange Trusted Subsystem', false, true, E'MATCH (n:Group)\nWHERE n.name STARTS WITH ''EXCHANGE TRUSTED SUBSYSTEM@''\nRETURN n;', E'The Exchange Trusted Subsystem group has takeover permissions on all users with the default ACL inheritance enabled from the domain, regardless of the permission model Exchange is configured to. The compromising permission is write access to the AltSecurityIdentities attribute, which allows an attacker to add an explicit mapping for the user for domain authentication. Typically, some Tier Zero users inherit permissions from the domain. The group is therefore Tier Zero.\n\nThe group can only be treated as non-Tier Zero if all Tier Zero users are protected from this compromising permission.'),
('Exchange Windows Permissions', false, true, E'MATCH (n:Group)\nWHERE n.name STARTS WITH ''EXCHANGE WINDOWS PERMISSIONS@''\nRETURN n;', E'The Exchange Windows Permissions group has takeover permissions on all users (WriteDACL and reset password) and all groups (edit membership) with the default ACL inheritance enabled from the domain, if Exchange is configured with the default shared permission model or the RBAC split model. Typically, some Tier Zero users and groups inherit permissions from the domain. The group is therefore Tier Zero.\n\nIf Exchange is configured in the AD split model, then this group has no compromising permissions and can be treated as non-Tier Zero.'),
('DNS Admins', false, true, E'MATCH (n:Group)\nWHERE n.name STARTS WITH ''DNSADMINS@''\nRETURN n;', E'DnsAdmins controls DNS which enables an attacker to trick a privileged victim to authenticate against an attacker-controlled host as it was another host. This enables a Kerberos relay attack. Also, control over DNS enables disruption of Tier Zero since Kerberos depends on DNS by default.\n\nThe group could previously use a feature in the Microsoft DNS management protocol to make the DNS service load any DLL and thereby obtain a session as SYSTEM on the DNS server. This vulnerability was patched in Dec 2021.'),
('Domain Controllers', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-516''\nRETURN n;', E'The Domain Controllers group has the GetChangesAll privilege on the domain. This is not enough to perform DCSync, where the GetChanges privilege is also required.\n\nThere are no known ways to abuse membership in this group to compromise Tier Zero. However, the GetChangesAll privilege is considered a security dependency that should only be held by Tier Zero principals. Additionally, control over the group allows one to impact the operability of Tier Zero by removing domain controllers from the group, which breaks AD replication. The group is therefore considered Tier Zero.'),
('Intune Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''3A2C62DB-5318-420D-8D74-23AFFEE5D9D5@''\nRETURN n;', E'The Intune Administrator role has permission to execute scripts locally on Entra-managed devices. The role has therefore a potential attack path to Tier Zero through Entra-managed devices used by Tier Zero principals. Furthermore, the Intune Administrator role can manage Conditional Access, which can be abused to lower the security of Tier Zero or prevent the operability of Tier Zero. The role is therefore considered Tier Zero.'),
('Security Administrator', false, true, E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''194AE4CB-B126-40B2-BD5B-6091B380977D@''\nRETURN n;', E'The Security Administrator role has access to Live Response API (if not disabled) with permission to execute scripts locally on Entra-managed devices. The role has therefore a potential attack path to Tier Zero through Entra-managed devices used by Tier Zero principals. Furthermore, the Security Administrator role can manage Conditional Access, which can be abused to lower the security of Tier Zero or prevent the operability of Tier Zero. The role is therefore considered Tier Zero.'),
('Cert Publishers', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-517''\nRETURN n;', E'The Cert Publishers group has full control permissions on root CA and AIA CA objects. This enables an attacker to add or remove certificates for these objects, which are trusted throughout the AD forest. As certificate authentication requires the certificate to chain up to a trusted root CA, an attacker could prevent successful authentication for AD accounts and disrupt Tier Zero operations. The group is therefore Tier Zero.\n\nIn some environments, the group also has full control over the NTAuth store. In that scenario, the group can take over the forest by adding a forged root certificate, making it trusted for NTAuth.'),
('NTAuth store', false, true, E'MATCH (n:NTAuthStore) \nRETURN n;', E'The NTAuth store is a security dependency for Tier Zero. A certificate that impersonates any user in AD must chain up to a trusted root CA and be issued by a CA trusted by the NTAuth store. With control over a root CA and the NTAuth store, an attacker can make an attacker-controlled root CA certificate meet these requirements and issue certificates as anyone, taking over Tier Zero. Control over the NTAuth store alone may be sufficient to disrupt Tier Zero operations, as the attacker can delete CA certificates that Tier Zero principals or systems rely on for authentication. The NTAuth store is therefore Tier Zero.'),
('Root CA object', false, true, E'MATCH (n:RootCA) \nRETURN n;', E'A root CA is a security dependency for Tier Zero. A certificate that impersonates any user in AD must chain up to a trusted root CA and be issued by a CA trusted by the NTAuth store. With control over a root CA and the NTAuth store, an attacker can make an attacker-controlled root CA certificate meet these requirements and issue certificates as anyone, taking over Tier Zero. Control over a root CA alone may be sufficient to disrupt Tier Zero operations, as the attacker can delete root CA certificates that Tier Zero principals or systems rely on for authentication. Root CA objects are therefore Tier Zero.'),
('Enterprise Key Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-527''\nRETURN n;', E'The Enterprise Key Admins group has write access to the msds-KeyCredentialLink attribute on all users (not protected by AdminSDHolder) and on all computers in the AD forest. This enables the group to compromise all these principals through Shadow Credentials attacks. The group is therefore considered Tier Zero.'),
('Key Admins', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-526''\nRETURN n;', E'The Key Admins group has write access to the msds-KeyCredentialLink attribute on all users (not protected by AdminSDHolder) and on all computers in the AD domain. This enables the group to compromise all these principals through Shadow Credentials attacks. The group is therefore considered Tier Zero.'),
('Cryptographic Operators', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-569''\nRETURN n;', E'The Cryptographic Operators group has the local privilege on domain controllers to perform cryptographic operations but no privilege to log in.\n\nThere are no known ways to abuse the membership of the group to compromise Tier Zero. The local privilege the group has on the domain controllers is considered security dependencies, and the group is therefore considered Tier Zero.'),
('Distributed COM Users', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-562''\nRETURN n;', E'The Distributed COM Users group has local privileges on domain controllers to launch, activate, and use Distributed COM objects but no privilege to log in.\n\nThere are no known ways to abuse the membership of the group to compromise Tier Zero. The local privileges the group has on the DCs are considered security dependency, and the group is therefore considered Tier Zero.'),
('AIA CA (AD object)', false, true, E'MATCH (n:AIACA) \nRETURN n;', E'The AIA CA objects may represent offline enterprise CAs or cross CAs. In such cases, deleting the AIA CA object would cause certificates, potentially of Tier Zero principals, to lose trust. We therefore recommend to treat AIACAs as Tier Zero.'),
('Enterprise CA (AD object)', false, true, E'MATCH (n:EnterpriseCA) \nRETURN n;', E'Control over an enterprise CA object enables an attacker to publish certificate templates. If any templates that allow ADCS domain escalation exist but are unpublished, then control over the enterprise CA object could enable a takeover of Tier Zero. An attacker could potentially also disrupt or takeover Tier Zero by deleting the certificate of the enterprise CA or changing the DNShostName of the enterprise CA to an attcker-controlled host. Enterprise CA objects are therfore Tier Zero.\n\nIf the enterprise CA certificate is removed from the NTAuth store, certificates from this CA cannot be used for domain authentication, thus preventing a Tier Zero takeover.'),
('Performance Log Users', false, true, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''S-1-5-32-559''\nRETURN n;', E'The Performance Log Users group has local privileges on domain controllers to launch, activate, and use Distributed COM objects but no privilege to log in.\n\nThere are no known ways to abuse the membership of the group to compromise Tier Zero. The local privileges the group has on the DCs are considered security dependency, and the group is therefore considered Tier Zero.'),
('Certificate template', false, true, E'MATCH (n:CertTemplate) \nRETURN n;', E'Control over a certificate template enables the ADCS ESC4 attack and Tier Zero takeover if the template is published to a CA trusted in the NTAuth store and that chains up to a trusted root CA. There are default templates that meet this requirement; others remain unpublished. A template cannot be used if it is not published, making control over an unpublished object less concerning. However, if it is ever published, it becomes a risk. We, therefore, recommend treating all certificate templates as Tier Zero objects, whether published or not.'),
('Azure tenant object', true, false, E'MATCH (n:AZTenant) \nRETURN n;', E'An attacker with control of the Tenant Root Object has control of all identities, applications, roles, and devices that reside in that tenant. Further, control of the Tenant Root Object enables an attacker to gain control of all Azure Resource Manager subscriptions that trust the tenant. This object is therefore considered Tier Zero.'),
('Enterprise Domain Controllers', true, false, E'MATCH (n:Group)\nWHERE n.objectid ENDS WITH ''-1-5-9''\nRETURN n;', E'There are no known ways to abuse membership in this group to compromise Tier Zero. However, the GetChangesAll privilege is considered a security dependency that should only be held by Tier Zero principals. Additionally, control over the group allows one to impact the operability of Tier Zero by removing domain controllers from the group, which breaks AD replication. The group is therefore considered Tier Zero."'),
('AZUREADSSOACC object', false, true, E'MATCH (n:Computer)\nWHERE n.samaccountname = ''AZUREADSSOACC$''\nRETURN n;', E'Microsoft automatically creates the AZUREADSSOACC account when enabling Seamless SSO. When configured for Seamless SSO, this object has the ability to modify any synced object within an Azure environment, granting significant control over the organization.')
-- END
	) AS s (name, enabled, allow_disable, cypher, description)
), inserted_selectors AS (
	INSERT INTO asset_group_tag_selectors (
		asset_group_tag_id,
		created_at,
		created_by,
		updated_at,
		updated_by,
		disabled_at,
		disabled_by,
		name,
		description,
		is_default,
		allow_disable,
		auto_certify
	)
	SELECT
		(SELECT id FROM asset_group_tags WHERE name = 'Tier Zero'),
		current_timestamp,
		'SYSTEM',
		current_timestamp,
		'SYSTEM',
		CASE WHEN NOT d.enabled THEN current_timestamp ELSE NULL END,
		CASE WHEN NOT d.enabled THEN 'SYSTEM' ELSE NULL END,
		d.name,
		d.description,
		true,
		d.allow_disable,
		false
	FROM src_data d WHERE NOT EXISTS(SELECT 1 FROM asset_group_tag_selectors WHERE name = d.name)
	RETURNING id, name
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
	s.id,
	2,
	d.cypher
FROM inserted_selectors s JOIN src_data d ON d.name = s.name;

-- Enable `back_button_support` feature flag and block users from updating it.
UPDATE feature_flags SET user_updatable = false, enabled = true WHERE key = 'back_button_support';

UPDATE feature_flags SET "user_updatable" = true WHERE "key" = 'tier_management_engine';
