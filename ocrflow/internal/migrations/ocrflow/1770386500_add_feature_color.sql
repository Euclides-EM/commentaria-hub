-- Add color property for features.
ALTER TABLE features ADD COLUMN color TEXT NOT NULL DEFAULT '#000000' CHECK (color <> '');

-- Default colors for title page features (dataset_id = 'tps').
UPDATE features SET color = '#FADADD' WHERE dataset_id = 'tps' AND (id = 'base_content' OR name = 'Base Content');
UPDATE features SET color = '#AEC6CF' WHERE dataset_id = 'tps' AND (id = 'base_content_description' OR name = 'Base Content Description');
UPDATE features SET color = '#909fd7' WHERE dataset_id = 'tps' AND (id = 'editor_name' OR name = 'Editor Name');
UPDATE features SET color = '#FFDAB9' WHERE dataset_id = 'tps' AND (id = 'editor_description' OR name = 'Editor Description');
UPDATE features SET color = '#D4C5F9' WHERE dataset_id = 'tps' AND (id = 'dedicatee_name' OR name = 'Dedicatee Name');
UPDATE features SET color = '#FFC1CC' WHERE dataset_id = 'tps' AND (id = 'edition_details' OR name = 'Edition Details');
UPDATE features SET color = '#9783d2' WHERE dataset_id = 'tps' AND (id = 'supplementary_content' OR name = 'Supplementary Content');
UPDATE features SET color = '#D1E7E0' WHERE dataset_id = 'tps' AND (id = 'printing_privilege' OR name = 'Printing Privilege');
UPDATE features SET color = '#F0E68C' WHERE dataset_id = 'tps' AND (id = 'references_to_euclid' OR name = 'References to Euclid');
UPDATE features SET color = '#e567ac' WHERE dataset_id = 'tps' AND (id = 'educational_authorities_references' OR name = 'Educational Authorities References');
UPDATE features SET color = '#e59c67' WHERE dataset_id = 'tps' AND (id = 'origin_language' OR name = 'Origin Language');
UPDATE features SET color = '#b0e57c' WHERE dataset_id = 'tps' AND (id = 'description_of_euclid' OR name = 'Description of Euclid');
UPDATE features SET color = '#954caf' WHERE dataset_id = 'tps' AND (id = 'action_verbs' OR name = 'Action Verbs');
UPDATE features SET color = '#E4A0D8' WHERE dataset_id = 'tps' AND (id = 'audience' OR name = 'Audience');
UPDATE features SET color = '#A3D5C3' WHERE dataset_id = 'tps' AND (id = 'elements_designation' OR name = 'Elements Designation');
UPDATE features SET color = '#F0B2A1' WHERE dataset_id = 'tps' AND (id = 'greek_text' OR name = 'Greek Text');
UPDATE features SET color = '#B0C4DE' WHERE dataset_id = 'tps' AND (id = 'institutions' OR name = 'Institutions');
UPDATE features SET color = '#FFB6C1' WHERE dataset_id = 'tps' AND (id = 'bound_with' OR name = 'Bound With');
UPDATE features SET color = '#D3D3D3' WHERE dataset_id = 'tps' AND (id = 'enriched_with' OR name = 'Enriched With');
UPDATE features SET color = '#FFDEAD' WHERE dataset_id = 'tps' AND (id = 'date_in_imprint' OR name = 'Date in Imprint');
UPDATE features SET color = '#ADD8E6' WHERE dataset_id = 'tps' AND (id = 'publisher_in_imprint' OR name = 'Publisher in Imprint');
UPDATE features SET color = '#E6E6FA' WHERE dataset_id = 'tps' AND (id = 'location_in_imprint' OR name = 'Location in Imprint');
UPDATE features SET color = '#D1E7E0' WHERE dataset_id = 'tps' AND (id = 'printing_privilege_in_imprint' OR name = 'Printing Privilege in Imprint');
UPDATE features SET color = '#D4C5F9' WHERE dataset_id = 'tps' AND (id = 'dedication_in_imprint' OR name = 'Dedication in Imprint');
UPDATE features SET color = '#909fd7' WHERE dataset_id = 'tps' AND (id = 'editor_in_imprint' OR name = 'Editor in Imprint');
UPDATE features SET color = '#FFDAB9' WHERE dataset_id = 'tps' AND (id = 'editor_description_in_imprint' OR name = 'Editor Description in Imprint');
