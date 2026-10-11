ALTER TABLE notification_incidents DROP CONSTRAINT notification_incidents_category_check;
ALTER TABLE notification_incidents ADD CHECK (category IN ('source_delay','processing_failed','backup_failed','processing_backlog'));
ALTER TABLE notification_incidents ADD COLUMN workload jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(workload)='object');
ALTER TABLE notification_messages ADD COLUMN workload jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(workload)='object');
