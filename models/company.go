package models

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type CompanyAPI struct {
	DB *sql.DB
}

type Company struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

const AddUserToCompanyQuery = `
	with find_user (
		SELECT * FROM USERS 
		WHERE Email=$2
	)	

	INSERT INTO USER_COMPANY(CompanyId,UserId)
	SELECT $1,Id
	FROM USERS 
	WHERE EXISTS (
		SELECT ID FROM USERS WHERE Email=$2
	) AND Email=$2
`

// AddUser adds a user by email address
func (c CompanyAPI) AddUser(ctx context.Context, db *sql.DB, companyID, userEmail string) (err error) {
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

	res, err := tx.ExecContext(ctx, AddUserToCompanyQuery, companyID, userEmail)
	if err != nil {
		return
	}

	nrows, err := res.RowsAffected()
	if err != nil || nrows == 0 {
		return errors.New("not able to add user")
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	return err
}

const searchCompanyQuery = `
	SELECT ID
	,Name
	FROM COMPANY
	WHERE Name like $1
`

// SearchCompany receives a string query and returns a list of company results
func (c CompanyAPI) SearchCompany(ctx context.Context, db *sql.DB, query string) (queryResult []Company, err error) {
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

	rows, err := tx.QueryContext(ctx, searchCompanyQuery, query)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var company Company
		if err = rows.Scan(&company); err != nil {
			return
		}
		queryResult = append(queryResult, company)
	}

	if err = rows.Close(); err != nil {
		return
	}

	if err = rows.Err(); err != nil {
		return
	}
	return
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

func (c Company) CreateTeam(ctx context.Context, db *sql.DB, team Team) (err error) {
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

	res, err := tx.ExecContext(ctx, createTeamQuery, team.Name)
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
func (c Company) AssignPersonToTeam(ctx context.Context, db *sql.DB) error {
	return nil
}
