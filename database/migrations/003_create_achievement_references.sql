-- 7. Achievement References
CREATE TYPE achievement_status AS ENUM ('draft', 'submitted', 'verified', 'rejected', 'deleted');

CREATE TABLE IF NOT EXISTS achievement_references (
    id UUID PRIMARY KEY,
    student_id UUID REFERENCES students(id),
    mongo_achievement_id VARCHAR(24) NOT NULL,
    status achievement_status DEFAULT 'draft',
    submitted_at TIMESTAMP,
    verified_at TIMESTAMP,
    verified_by UUID REFERENCES users(id),
    rejection_note TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
