ALTER TABLE observation_campaigns ADD COLUMN scope text NOT NULL DEFAULT 'mvp'
 CHECK (scope IN ('mvp','municipality'));
