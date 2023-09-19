-- Copyright 2023 Specter Ops, Inc.
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

-- FIRST, REMOVE DUPLICATE RECORDS WITH BOTH NAME AND SELECTOR MATCHING
DELETE FROM asset_group_selectors ag1
USING asset_group_selectors ag2
WHERE ag1.id < ag2.id
      AND ag1.name = ag2.name
      AND ag1.selector = ag2.selector;

-- IF ONLY NAME MATCHES BUT SELECTOR DOES NOT, THEN APPEND _2, _3 ETC TO THE NAME COLUMN
DO $$
	DECLARE
    seq integer := 2;
	  prev integer := 1;
  BEGIN
  WHILE EXISTS(
    SELECT 1
    FROM asset_group_selectors
    GROUP BY name
    HAVING COUNT(*)>1
  ) LOOP
      IF seq < 3 THEN
        UPDATE asset_group_selectors ag1
        SET name = ag1.name || '_' || seq
        FROM asset_group_selectors ag2
        WHERE ag1.id < ag2.id
              AND ag1.name = ag2.name;
      ELSE
        UPDATE asset_group_selectors ag1
        SET name = TRIM(TRAILING '_' || prev FROM ag1.name) || '_' || seq
        FROM asset_group_selectors ag2
        WHERE ag1.id < ag2.id
              AND ag1.name = ag2.name;
      END IF;
      seq := seq + 1;
      prev := prev + 1;
  END LOOP;
END$$;

-- NOW THAT ALL DUPLICATE NAMES HAVE BEEN REMOVED, ADD THE UNIQUENESS CONSTRAINT ON THE NAME COLUMN
ALTER TABLE asset_group_selectors
ADD CONSTRAINT asset_group_selectors_unique_name UNIQUE (name);
