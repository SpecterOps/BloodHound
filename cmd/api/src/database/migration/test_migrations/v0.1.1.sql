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

-- Name should now be unique, so rename duplicates and enable constraint
-- FIRST, REMOVE DUPLICATE RECORDS WITH BOTH NAME AND FOO MATCHING
DELETE FROM migration_test mt1 USING migration_test mt2
WHERE mt1.id < mt2.id
  AND mt1.name = mt2.name
  AND mt1.foo = mt2.foo;

-- IF ONLY NAME MATCHES BUT FOO DOES NOT, THEN APPEND _2, _3 ETC TO THE NAME COLUMN
DO $$
  DECLARE
    seq integer := 2;
    prev integer := 1;
  BEGIN
  WHILE EXISTS(
    SELECT 1
    FROM migration_test
    GROUP BY name
    HAVING COUNT(*)>1
  ) LOOP
      IF seq < 3 THEN
        UPDATE migration_test mt1
        SET name = mt1.name || '_' || seq
        FROM migration_test mt2
        WHERE mt1.id < mt2.id
              AND mt1.name = mt2.name;
      ELSE
        UPDATE migration_test mt1
        SET name = TRIM(TRAILING '_' || prev FROM mt1.name) || '_' || seq
        FROM migration_test mt2
        WHERE mt1.id < mt2.id
              AND mt1.name = mt2.name;
      END IF;
      seq := seq + 1;
      prev := prev + 1;
  END LOOP;
END$$;

-- NOW THAT ALL DUPLICATE NAMES HAVE BEEN REMOVED, ADD THE UNIQUENESS CONSTRAINT ON THE NAME COLUMN
ALTER TABLE migration_test
ADD CONSTRAINT migration_test_unique_name UNIQUE (name);
