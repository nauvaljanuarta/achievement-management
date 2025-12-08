package database

import (
    "database/sql"
    "log"
    "os"
    "path/filepath"
)

func Migrate(db *sql.DB, path string) {
    files, err := filepath.Glob(path + "/*.sql")
    if err != nil {
        log.Fatal(err)
    }

    for _, file := range files {
        sqlBytes, err := os.ReadFile(file) 
        if err != nil {
            log.Fatalf("Failed to read %s: %v", file, err)
        }

        _, err = db.Exec(string(sqlBytes))
        if err != nil {
            log.Fatalf("Failed to execute %s: %v", file, err)
        }

        log.Printf("Migrated: %s", file)
    }
}
