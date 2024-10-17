CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule DATETIME,                    
    user_schedule TEXT,                    
    occurrences INTEGER NOT NULL,                   
    frequency TEXT NOT NULL,
    label TEXT NOT NULL,
    cron_time TEXT,               
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT,                           
    action TEXT,                           
    args TEXT,                             
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);
