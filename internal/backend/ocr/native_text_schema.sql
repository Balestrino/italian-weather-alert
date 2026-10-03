-- Preserve existing image-page records while representing locally extracted PDF
-- text without falsely labelling its input as a rendered image.
ALTER TABLE ocr_page_results DROP CONSTRAINT ocr_page_results_media_type_check;
ALTER TABLE ocr_page_results ADD CONSTRAINT ocr_page_results_media_type_check
 CHECK (media_type IN ('image/png','image/jpeg','application/pdf'));
