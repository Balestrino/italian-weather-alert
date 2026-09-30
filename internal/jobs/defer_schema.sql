ALTER TABLE processing_attempts DROP CONSTRAINT processing_attempts_outcome_check;
ALTER TABLE processing_attempts ADD CHECK (outcome IN ('running','succeeded','retry','failed','abandoned','deferred'));
