package test

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const dbURL = "postgres://localhost:5400/db?user=db&password=Pwd1&sslmode=disable"

func NewTestDB() (*sql.DB, error) {

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	m, err := migrate.New("file://../db/migrations", dbURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, errors.WithStack(err)
	}

	return db, nil
}

func TestMigrations(db *sql.DB) error {
	m, err := migrate.New("file://../testsql", dbURL)
	if err != nil {
		return errors.WithStack(err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.WithStack(err)
	}
	return nil
}

const emptyQuery = `
	DROP SCHEMA public CASCADE;
	CREATE SCHEMA public;
`

func EmptyTestDB(db *sql.DB) {
	ctx := context.Background()
	res, err := db.ExecContext(ctx, emptyQuery)
	if err != nil {
		fmt.Println("error ", err.Error())
	}

	nrows, err := res.RowsAffected()
	if err != nil {
		fmt.Println("error ", err.Error())
	}

	if nrows == 0 {
		fmt.Println("no rows affected when emptying db")
	}
	return
}
