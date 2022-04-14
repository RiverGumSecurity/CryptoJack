package cjlib

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "os"
)

func CreateAndConnectDB(sqlfilename string) *sql.DB {
    if _, err := os.Stat(sqlfilename); err == nil {
        os.Remove(sqlfilename)
    }
    file, err := os.Create(sqlfilename)
    if err != nil {
        log.Fatal(err.Error())
    }
    file.Close()
    return ConnectDB(sqlfilename)
}

func ConnectDB(sqlfilename string) *sql.DB {
    dbh, _ := sql.Open("sqlite3", sqlfilename)
    createTable(dbh)
    return dbh
}

func createTable(db *sql.DB) {
    sql := `CREATE TABLE IF NOT EXISTS filehashes (
        "path" TEXT NOT NULL PRIMARY KEY,
        "sha256hash" TEXT);`
    st, err := db.Prepare(sql)
    if err != nil {
        log.Fatal(err.Error())
    }
    st.Exec()
}

func InsertFilePathHash(db *sql.DB, path string, sha256hash string) {
    sql := `INSERT INTO filehashes(path, sha256hash) VALUES (?, ?)`
    st, err := db.Prepare(sql)
    if err != nil {
        log.Fatalln(err.Error())
    }
    _, err = st.Exec(path, sha256hash)
    if err != nil {
        log.Fatalln(err.Error())
    }
}

func GetFilePathHash(db *sql.DB, path string) string {
    sql := `SELECT sha256hash FROM filehashes WHERE path = ?`
    row, err := db.Query(sql, path)
    if err != nil {
        log.Fatal(err)
    }
    var hash string
    row.Next()
    row.Scan(&hash)
    row.Close()
    return hash
}
