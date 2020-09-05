package models

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type Company struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const createCompanyQuery = `
	INSERT INTO COMPANY(Name)
	VALUES ($1)
	RETURNING ID
`

func (c Company) CreateCompany(ctx context.Context, db *sql.DB) (companyID int64, err error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	res, err := tx.ExecContext(ctx, createCompanyQuery, c.Name)
	if err != nil {
		return
	}

	companyID, err = res.LastInsertId()
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	return companyID, err
}

const createTeamQuery = `
	INSERT INTO TEAM(Name)
	VALUES ($1)
`

func (t Team) CreateTeam(ctx context.Context, db *sql.DB) (err error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	res, err := tx.ExecContext(ctx, createTeamQuery, t.Name)
	if err != nil {
		return
	}

	if rowsAffected, err := res.RowsAffected(); err != nil || rowsAffected == 0 {
		return errors.New("not able to update teams table")
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	return nil
}

func (c Company) AssignPersonToCompany(ctx context.Context, db *sql.DB) error {
	return nil
}

// AssignUserToTeam assigns a userID to a teamID as long as that
// user is a part om the mother company
func (t Team) AssignPersonToTeam(ctx context.Context, db *sql.DB) error {
	return nil
}
