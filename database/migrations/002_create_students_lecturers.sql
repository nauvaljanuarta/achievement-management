-- 5. Lecturers
CREATE TABLE IF NOT EXISTS lecturers (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    lecturer_id VARCHAR(20) UNIQUE NOT NULL,
    department VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 6. Students
CREATE TABLE IF NOT EXISTS students (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    student_id VARCHAR(20) UNIQUE NOT NULL,
    program_study VARCHAR(100),
    academic_year VARCHAR(10),
    advisor_id UUID REFERENCES lecturers(id),
    created_at TIMESTAMP DEFAULT NOW()
);
