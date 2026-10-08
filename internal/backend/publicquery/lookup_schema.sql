-- Chronological version selection and newer-version checks.
CREATE INDEX retained_versions_public_chronology ON retained_versions(document_id,first_acquired_at DESC,id DESC);
CREATE INDEX retained_acquisitions_public_visibility ON retained_acquisitions(version_id,document_id,source_id,configuration);

-- Select matching receipts before evaluating their visibility and field evidence.
CREATE INDEX verification_public_candidate ON domain_verification_receipts(candidate_version_id);
CREATE INDEX verification_public_primary_record ON domain_verification_receipts(kind,(body->>'primary_record_id'));
CREATE INDEX verification_public_domain_record ON domain_verification_receipts(kind,(body->>'domain_record_id'));
CREATE INDEX verification_public_evidence ON domain_verification_dependencies(evidence_version_id,receipt_id);
