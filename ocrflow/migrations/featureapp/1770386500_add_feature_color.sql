-- Add color property for features.
ALTER TABLE features ADD COLUMN color TEXT NOT NULL DEFAULT '#000000' CHECK (color <> '');

-- Default colors for title page features (collection_id = 'tps').
UPDATE features SET color = '#FADADD' WHERE collection_id = 'tps' AND (id = 'base_content' OR name = 'Base Content');
UPDATE features SET color = '#AEC6CF' WHERE collection_id = 'tps' AND (id = 'base_content_description' OR name = 'Base Content Description');
UPDATE features SET color = '#909fd7' WHERE collection_id = 'tps' AND (id = 'editor_name' OR name = 'Adapter Attribution');
UPDATE features SET color = '#FFDAB9' WHERE collection_id = 'tps' AND (id = 'editor_description' OR name = 'Adapter Description');
UPDATE features SET color = '#D4C5F9' WHERE collection_id = 'tps' AND (id = 'dedicatee_name' OR name = 'Patronage Dedication');
UPDATE features SET color = '#FFC1CC' WHERE collection_id = 'tps' AND (id = 'edition_details' OR name = 'Edition Statement');
UPDATE features SET color = '#9783d2' WHERE collection_id = 'tps' AND (id = 'supplementary_content' OR name = 'Supplementary Content');
UPDATE features SET color = '#D1E7E0' WHERE collection_id = 'tps' AND (id = 'printing_privilege' OR name = 'Publishing Privileges');
UPDATE features SET color = '#F0E68C' WHERE collection_id = 'tps' AND (id = 'references_to_Euclid' OR name = 'Euclid References');
UPDATE features SET color = '#e567ac' WHERE collection_id = 'tps' AND (id = 'educational_authorities_references' OR name = 'Other Educational Authorities');
UPDATE features SET color = '#e59c67' WHERE collection_id = 'tps' AND (id = 'origin_language' OR name = 'Explicit Language References');
UPDATE features SET color = '#b0e57c' WHERE collection_id = 'tps' AND (id = 'description_of_Euclid' OR name = 'Euclid Description');
UPDATE features SET color = '#954caf' WHERE collection_id = 'tps' AND (id = 'action_verbs' OR name = 'Verbs');
UPDATE features SET color = '#E4A0D8' WHERE collection_id = 'tps' AND (id = 'audience' OR name = 'Intended Audience');
UPDATE features SET color = '#A3D5C3' WHERE collection_id = 'tps' AND (id = 'Elements_designation' OR name = 'Elements Designation');
UPDATE features SET color = '#F0B2A1' WHERE collection_id = 'tps' AND (id = 'Greek_text' OR name = 'Greek designation');
UPDATE features SET color = '#B0C4DE' WHERE collection_id = 'tps' AND (id = 'institutions' OR name = 'Institutions');
UPDATE features SET color = '#FFB6C1' WHERE collection_id = 'tps' AND (id = 'bound_with' OR name = 'Bound With');
UPDATE features SET color = '#D3D3D3' WHERE collection_id = 'tps' AND (id = 'enriched_with' OR name = 'Enriched With');
UPDATE features SET color = '#FFDEAD' WHERE collection_id = 'tps' AND (id = 'date_in_imprint' OR name = 'Date in Imprint');
UPDATE features SET color = '#ADD8E6' WHERE collection_id = 'tps' AND (id = 'publisher_in_imprint' OR name = 'Publisher in Imprint');
UPDATE features SET color = '#E6E6FA' WHERE collection_id = 'tps' AND (id = 'location_in_imprint' OR name = 'Place in Imprint');
UPDATE features SET color = '#D1E7E0' WHERE collection_id = 'tps' AND (id = 'printing_privilege_in_imprint' OR name = 'Privileges in Imprint');
UPDATE features SET color = '#D4C5F9' WHERE collection_id = 'tps' AND (id = 'dedication_in_imprint' OR name = 'Dedication in Imprint');
UPDATE features SET color = '#909fd7' WHERE collection_id = 'tps' AND (id = 'editor_in_imprint' OR name = 'Adapter Attribution in Imprint');
UPDATE features SET color = '#FFDAB9' WHERE collection_id = 'tps' AND (id = 'editor_description_in_imprint' OR name = 'Adapter Description in Imprint');
