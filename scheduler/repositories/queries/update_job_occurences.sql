UPDATE jobs
SET occurrences = occurrences - 1
WHERE id = ?
RETURNING occurrences;