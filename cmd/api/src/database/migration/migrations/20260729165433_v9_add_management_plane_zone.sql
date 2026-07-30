-- Copyright 2026 Specter Ops, Inc.
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

-- +goose Up

INSERT INTO kind (name)
VALUES ('Tag_Management_Plane')
ON CONFLICT (name) DO NOTHING;

UPDATE asset_group_tags
SET position = position + 1
WHERE type = 1
  AND position >= 2
  AND NOT EXISTS (
    SELECT 1
    FROM asset_group_tags
    WHERE name = 'Management Plane'
      AND deleted_at IS NULL
  );

INSERT INTO asset_group_tags (
    type,
    kind_id,
    name,
    description,
    created_at,
    created_by,
    updated_at,
    updated_by,
    position,
    require_certify,
    analysis_enabled
)
SELECT
    1,
    kind.id,
    'Management Plane',
    'The management plane is the layer designated for enterprise-wide IT management functions, managing workloads and infrastructure, and controlling configuration and operational tasks across systems.',
    current_timestamp,
    'SYSTEM',
    current_timestamp,
    'SYSTEM',
    2,
    false,
    false
FROM kind
WHERE kind.name = 'Tag_Management_Plane'
  AND NOT EXISTS (
    SELECT 1
    FROM asset_group_tags
    WHERE name = 'Management Plane'
      AND deleted_at IS NULL
  );

-- Role names, descriptions, and template IDs are sourced from:
-- https://learn.microsoft.com/entra/identity/role-based-access-control/permissions-reference
WITH source_data AS (
    SELECT *
    FROM (VALUES
        ('AI Administrator', 'Manage all aspects of Microsoft 365 Copilot and AI-related enterprise services in Microsoft 365.', 'D2562EDE-74DB-457E-A7B6-544E236EBB61'),
        ('Attribute Assignment Administrator', 'Assign custom security attribute keys and values to supported Microsoft Entra objects.', '58A13EA3-C632-46AE-9EE0-9C0D43CD7F3D'),
        ('Attribute Provisioning Administrator', 'Read and edit the provisioning configuration of all active custom security attributes for an application.', 'ECB2C6BF-0AB6-418E-BD87-7986F8D63BBE'),
        ('Authentication Administrator', 'Can access to view, set and reset authentication method information for any non-admin user.', 'C4E39BD9-1100-46D3-8C65-FB160DA0071F'),
        ('Authentication Extensibility Administrator', 'Customize sign in and sign up experiences for users by creating and managing custom authentication extensions.', '25A516ED-2FA0-40EA-A2D0-12923A21473A'),
        ('Authentication Policy Administrator', 'Can create and manage the authentication methods policy, tenant-wide MFA settings, password protection policy, and verifiable credentials.', '0526716B-113D-4C15-B2C8-68E3C22B9F80'),
        ('B2C IEF Keyset Administrator', 'Can manage secrets for federation and encryption in the Identity Experience Framework (IEF).', 'AAF43236-0C0D-4D5F-883A-6955382AC081'),
        ('Cloud App Security Administrator', 'Manage all aspects of the Defender for Cloud Apps product.', '892C5842-A9A6-463A-8041-72AA08CA3CF6'),
        ('Compliance Administrator', 'Can read and manage compliance configuration and reports in Microsoft Entra ID and Microsoft 365.', '17315797-102D-40B4-93E0-432062CACA18'),
        ('Directory Synchronization Accounts', 'Only used by Microsoft Entra Connect service.', 'D29B2B05-8046-44BA-8758-1E26182FCF32'),
        ('Directory Writers', 'Can read and write basic directory information. For granting access to applications, not intended for users.', '9360FEB5-F418-4BAA-8175-E2A00BAC4301'),
        ('Domain Name Administrator', 'Can manage domain names in cloud and on-premises.', '8329153B-31D0-4727-B945-745EB3BC5F31'),
        ('Dynamics 365 Administrator', 'Can manage all aspects of the Dynamics 365 product.', '44367163-EBA1-44C3-98AF-F5787879F96A'),
        ('Exchange Administrator', 'Can manage all aspects of the Exchange product.', '29232CDF-9323-42FD-ADE2-1D097AF3E4DE'),
        ('External ID User Flow Administrator', 'Can create and manage all aspects of user flows.', '6E591065-9BAD-43ED-90F3-E9424366D2F0'),
        ('External Identity Provider Administrator', 'Can configure identity providers for use in direct federation.', 'BE2F45A1-457D-42AF-A067-6EC1FA63BC45'),
        ('Global Secure Access Administrator', 'Create and manage all aspects of Global Secure Internet Access and Microsoft Global Secure Private Access, including managing access to public and private endpoints.', 'AC434307-12B9-4FA1-A708-88BF58CAABC1'),
        ('Groups Administrator', 'Members of this role can create/manage groups, create/manage groups settings like naming and expiration policies, and view groups activity and audit reports.', 'FDD7A751-B60B-444A-984C-02652FE8FA1C'),
        ('Helpdesk Administrator', 'Can reset passwords for non-administrators and Helpdesk Administrators.', '729827E3-9C14-49F7-BB1B-9608F156BBB8'),
        ('Identity Governance Administrator', 'Manage access using Microsoft Entra ID for identity governance scenarios.', '45D8D3C5-C802-45C6-B32A-1D70B5E1E86E'),
        ('Intune Administrator', 'Can manage all aspects of the Intune product.', '3A2C62DB-5318-420D-8D74-23AFFEE5D9D5'),
        ('Knowledge Administrator', 'Can configure knowledge, learning, and other intelligent features.', 'B5A8DCF3-09D5-43A9-A639-8E29EF291470'),
        ('Lifecycle Workflows Administrator', 'Create and manage all aspects of workflows and tasks associated with Lifecycle Workflows in Microsoft Entra ID.', '59D46F88-662B-457B-BCEB-5C3809E5908F'),
        ('Microsoft 365 Backup Administrator', 'Back up and restore content across supported services (SharePoint, OneDrive, and Exchange Online) in Microsoft 365 Backup', '1707125E-0AA2-4D4D-8655-A7C786C76A25'),
        ('Microsoft 365 Migration Administrator', 'Perform all migration functionality to migrate content to Microsoft 365 using Migration Manager.', '8C8B803F-96E1-4129-9349-20738D9F9652'),
        ('Partner Tier1 Support', 'Do not use - not intended for general use.', '4BA39CA4-527C-499A-B93D-D9B492C50246'),
        ('Password Administrator', 'Can reset passwords for non-administrators and Password Administrators.', '966707D0-3269-4727-9BE2-8C3A10F19B9D'),
        ('Power Platform Administrator', 'Manage all aspects of Microsoft Dynamics 365, Power Apps and Power Automate.', '11648597-926C-4CF3-9C36-BCEBB0BA8DCC'),
        ('Security Operator', 'Creates and manages security events and performs identity containment actions during security incidents.', '5F2222B1-57C3-48BA-8AD5-D4759F1FDE6F'),
        ('SharePoint Administrator', 'Can manage all aspects of the SharePoint service.', 'F28A1F50-F6E7-4571-818B-6A12F2AF6B6C'),
        ('Skype for Business Administrator', 'Can manage all aspects of the Skype for Business product.', '75941009-915A-4869-ABE7-691BFF18279E'),
        ('Teams Administrator', 'Can manage the Microsoft Teams service.', '69091246-20E8-4A56-AA4D-066075B2A7A8'),
        ('Teams Telephony Administrator', 'Manage voice and telephony features and troubleshoot communication issues within the Microsoft Teams service.', 'AA38014F-0993-46E9-9B45-30501A20909D'),
        ('User Administrator', 'Can manage all aspects of users and groups, including resetting passwords for limited admins.', 'FE930BE7-5E62-47DB-91AF-98C3A49A38B1'),
        ('Windows 365 Administrator', 'Can provision and manage all aspects of Cloud PCs.', '11451D60-ACB2-45EB-A7D6-43D0F0125C13'),
        ('Yammer Administrator', 'Manage all aspects of the Yammer service.', '810A2642-A034-447F-A5E8-41BEAA378541')
    ) AS roles (name, description, template_id)
),
inserted_selectors AS (
    INSERT INTO asset_group_tag_selectors (
        asset_group_tag_id,
        created_at,
        created_by,
        updated_at,
        updated_by,
        name,
        description,
        is_default,
        allow_disable,
        auto_certify
    )
    SELECT
        tag.id,
        current_timestamp,
        'SYSTEM',
        current_timestamp,
        'SYSTEM',
        role.name,
        role.description,
        true,
        true,
        0
    FROM source_data role
    CROSS JOIN (
        SELECT id
        FROM asset_group_tags
        WHERE name = 'Management Plane'
          AND deleted_at IS NULL
    ) tag
    WHERE NOT EXISTS (
        SELECT 1
        FROM asset_group_tag_selectors selector
        WHERE selector.asset_group_tag_id = tag.id
          AND selector.name = role.name
          AND selector.is_default = true
    )
    RETURNING id, name
)
INSERT INTO asset_group_tag_selector_seeds (selector_id, type, value)
SELECT
    selector.id,
    2,
    format(
        E'MATCH (n:AZRole) \nWHERE n.objectid STARTS WITH ''%s@''\nRETURN n;',
        role.template_id
    )
FROM inserted_selectors selector
JOIN source_data role ON role.name = selector.name;

-- +goose Down

DELETE FROM asset_group_tag_selectors
WHERE asset_group_tag_id = (
    SELECT id
    FROM asset_group_tags
    WHERE name = 'Management Plane'
      AND deleted_at IS NULL
)
  AND is_default = true
  AND name IN (
    'AI Administrator',
    'Attribute Assignment Administrator',
    'Attribute Provisioning Administrator',
    'Authentication Administrator',
    'Authentication Extensibility Administrator',
    'Authentication Policy Administrator',
    'B2C IEF Keyset Administrator',
    'Cloud App Security Administrator',
    'Compliance Administrator',
    'Directory Synchronization Accounts',
    'Directory Writers',
    'Domain Name Administrator',
    'Dynamics 365 Administrator',
    'Exchange Administrator',
    'External ID User Flow Administrator',
    'External Identity Provider Administrator',
    'Global Secure Access Administrator',
    'Groups Administrator',
    'Helpdesk Administrator',
    'Identity Governance Administrator',
    'Intune Administrator',
    'Knowledge Administrator',
    'Lifecycle Workflows Administrator',
    'Microsoft 365 Backup Administrator',
    'Microsoft 365 Migration Administrator',
    'Partner Tier1 Support',
    'Password Administrator',
    'Power Platform Administrator',
    'Security Operator',
    'SharePoint Administrator',
    'Skype for Business Administrator',
    'Teams Administrator',
    'Teams Telephony Administrator',
    'User Administrator',
    'Windows 365 Administrator',
    'Yammer Administrator'
  );

WITH deleted_zone AS (
    DELETE FROM asset_group_tags
    WHERE name = 'Management Plane'
      AND created_by = 'SYSTEM'
    RETURNING position
)
UPDATE asset_group_tags
SET position = position - 1
WHERE type = 1
  AND position > 2
  AND EXISTS (SELECT 1 FROM deleted_zone);

DELETE FROM kind
WHERE name = 'Tag_Management_Plane'
  AND NOT EXISTS (
    SELECT 1
    FROM asset_group_tags
    WHERE kind_id = kind.id
  );
