SELECT
    j.id,
    j.schedule,
    j.user_schedule,
    j.occurrences,
    j.frequency,
    j.label,
    j.created_at,
    
    w.action,
    w.args
FROM jobs j
LEFT JOIN workflows w ON j.id = w.job_id;
