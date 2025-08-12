package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Driver do SQLite
)

var DB *sql.DB

// InitDB inicializa a conexão com o banco de dados e cria as tabelas se não existirem.
func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./fruit_buckets.db")
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return migrate()
}

func InitDBTest() (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	if err := migrate(); err != nil {
		return nil, err
	}

	return DB, nil
}

func migrate() error {
	createTablesSQL := `
    CREATE TABLE IF NOT EXISTS buckets (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        capacity INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS fruits (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        price REAL NOT NULL,
        expiration_time INTEGER NOT NULL,
        bucket_id INTEGER,
        FOREIGN KEY(bucket_id) REFERENCES buckets(id) ON DELETE SET NULL
    );
    `

	_, err := DB.Exec(createTablesSQL)

	return err
}
