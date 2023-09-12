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
