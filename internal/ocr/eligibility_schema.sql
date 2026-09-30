ALTER TABLE ocr_resource_results DROP CONSTRAINT ocr_resource_results_status_check;
ALTER TABLE ocr_resource_results ADD CONSTRAINT ocr_resource_results_status_check CHECK(status IN('complete','partial_unreadable','unreadable','missing','skipped'));
