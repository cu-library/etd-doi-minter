USE drupal;
BEGIN;
SELECT
node.title,
field_data_dcterms_creator.dcterms_creator_value,
field_data_dcterms_date.dcterms_date_value,
IFNULL(field_data_thesis_degree_name.thesis_degree_name_first, ""),
IFNULL(GROUP_CONCAT(field_data_dcterms_identifier.dcterms_identifier_url SEPARATOR '|'), ""),
node.uuid
FROM node
LEFT JOIN field_data_dcterms_creator ON node.nid = field_data_dcterms_creator.entity_id
LEFT JOIN field_data_thesis_degree_name ON node.nid = field_data_thesis_degree_name.entity_id
LEFT JOIN field_data_dcterms_date ON node.nid = field_data_dcterms_date.entity_id
LEFT JOIN field_data_dcterms_identifier ON node.nid = field_data_dcterms_identifier.entity_id
WHERE node.type = 'etd' AND node.status = 1
GROUP BY node.nid
INTO OUTFILE '/tmp/etd-output.csv'
FIELDS TERMINATED BY ','
ENCLOSED BY '"'
ESCAPED BY '"'
LINES TERMINATED BY '\n';
ROLLBACK;