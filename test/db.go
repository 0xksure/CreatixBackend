package test

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
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
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	parent := filepath.Dir(wd)

	testsqlDir := fmt.Sprintf("/%s/%s", parent, "test/testsql")
	fmt.Println("testsqldir: ", testsqlDir)
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	mFiles, err := ioutil.ReadDir(testsqlDir)
	if err != nil {
		return err
	}

	for _, mFile := range mFiles {
		content, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", testsqlDir, mFile.Name()))
		if err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, string(content))
		if err != nil {
			return err
		}
		nrows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if nrows == 0 {
			return errors.New("no rows affected")
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

const emptyQuery = `
	DROP SCHEMA public CASCADE;
	CREATE SCHEMA public;
`

func EmptyTestDB(t *testing.T, db *sql.DB) {
	t.Log("Empty db")
	defer db.Close()
	res, err := db.Exec(emptyQuery)
	require.NoError(t, err, "not able to empty db")

	nrows, err := res.RowsAffected()
	require.NoError(t, err, "not able to empty db")
	require.NotEqual(t, 0, nrows)

}
