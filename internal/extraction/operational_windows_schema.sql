ALTER TABLE extraction_segments
 ADD COLUMN core_start_byte integer,
 ADD COLUMN core_end_byte integer,
 ADD COLUMN source_start_byte integer,
 ADD COLUMN source_end_byte integer,
 ADD COLUMN normalization_map jsonb NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE extraction_segments
 ADD CONSTRAINT extraction_segments_core_bounds CHECK (
  (core_start_byte IS NULL AND core_end_byte IS NULL) OR
  (core_start_byte >= start_byte AND core_end_byte > core_start_byte AND core_end_byte <= end_byte)
 ),
 ADD CONSTRAINT extraction_segments_source_bounds CHECK (
  (source_start_byte IS NULL AND source_end_byte IS NULL) OR
  (source_start_byte >= 0 AND source_end_byte > source_start_byte)
 ),
 ADD CONSTRAINT extraction_segments_normalization_map_array CHECK (jsonb_typeof(normalization_map) = 'array');
