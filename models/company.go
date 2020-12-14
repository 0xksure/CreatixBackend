package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/pkg/errors"
)

type Company struct {
	ID   string `json:"id"`
	Name string `json:"companyName"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type CompanyClient interface {
	// Company
	CreateCompany(ctx context.Context, company, userId string) (companyID *int64, err error)
	AddUserToCompanyByEmail(ctx context.Context, companyID, userEmail string) (err error)
	SearchCompany(ctx context.Context, query string) (queryResult []Company, err error)
	GetCompaniesAssociatedToUser(ctx context.Context, userID string) (companies []Company, err error)

	// Team
	CreateTeam(ctx context.Context, team Team) (err error)
	AddUserToTeam(ctx context.Context) error
}

type companyClient struct {
	DB *sql.DB
}

// NewCompanyClient creates a new company client
func NewCompanyClient(DB *sql.DB) CompanyClient {
	return companyClient{DB: DB}
}

const getCompaniesAssociatedToUserQuery = `
	SELECT 
	c.Id, 
	c.Name
	FROM COMPANY c
	LEFT JOIN (
		SELECT CompanyId
		FROM USER_COMPANY 
		WHERE UserId=$1
	) as uc 
	ON c.Id=uc.CompanyId
`

func (c companyClient) GetCompaniesAssociatedToUser(ctx context.Context, userID string) (companies []Company, err error) {
	rows, err := c.DB.QueryContext(ctx, getCompaniesAssociatedToUserQuery, userID)
	if err != nil {
		return companies, err
	}
	defer rows.Close()

	for rows.Next() {
		var company Company
		if err = rows.Scan(&company.ID, &company.Name); err != nil {
			return
		}
		companies = append(companies, company)
	}

	if rows.Close() != nil {
		return
	}

	if rows.Err() != nil {
		return
	}

	return
}

const addUserToCompanyByEmailQuery = `
	INSERT INTO USER_COMPANY(CompanyId,UserId)
	VALUES ($1,$2)
`

// AddUser adds a user by email address
func (c companyClient) AddUserToCompanyByEmail(ctx context.Context, companyID, userEmail string) (err error) {

	user, err := utils.FindUserByEmail(ctx, c.DB, userEmail)
	if err != nil {
		return
	}

	res, err := c.DB.ExecContext(ctx, addUserToCompanyByEmailQuery, companyID, user.ID)
	if err != nil {
		return
	}

	nrows, err := res.RowsAffected()
	if err != nil || nrows == 0 {
		return errors.New("not able to add user")
	}
	fmt.Println("Number of rows affected: ", nrows)
	return nil
}

const searchCompanyQuery = `
	SELECT ID
	,Name
	FROM COMPANY
	WHERE Name like $1
`

// SearchCompany receives a string query and returns a list of company results
func (c companyClient) SearchCompany(ctx context.Context, query string) (queryResult []Company, err error) {
	rows, err := c.DB.QueryContext(ctx, searchCompanyQuery, query)
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
const addUserToCompanyByID = `
	INSERT INTO USER_COMPANY(CompanyId,UserId)
	VALUES ($1,$2)
`

// CreateCompany creates a new company and adds the current user to it
func (c companyClient) CreateCompany(ctx context.Context, companyName, userID string) (companyID *int64, err error) {
	tx, err := c.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	err = tx.QueryRowContext(ctx, createCompanyQuery, companyName).Scan(&companyID)
	if err != nil {
		return
	}

	if companyID == nil {
		return
	}

	res, err := tx.ExecContext(ctx, addUserToCompanyByID, companyID, userID)
	if err != nil {
		return
	}
	nrows, err := res.RowsAffected()
	if err != nil || nrows == 0 {
		return companyID, errors.New("no company added")
	}

	if err = tx.Commit(); err != nil {
		return companyID, err
	}

	return companyID, err
}

const createTeamQuery = `
	INSERT INTO TEAM(Name)
	VALUES ($1)
`

func (c companyClient) CreateTeam(ctx context.Context, team Team) (err error) {

	res, err := c.DB.ExecContext(ctx, createTeamQuery, team.Name)
	if err != nil {
		return
	}

	if rowsAffected, err := res.RowsAffected(); err != nil || rowsAffected == 0 {
		return errors.New("not able to update teams table")
	}

	return nil
}

// AssignUserToTeam assigns a userID to a teamID as long as that
// user is a part om the mother company
func (c companyClient) AddUserToTeam(ctx context.Context) error {
	return nil
}
