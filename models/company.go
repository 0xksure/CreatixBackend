package models

import (
	"context"
	"database/sql"

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
type CompanyClienter interface {
	// Company
	CreateCompany(ctx context.Context, company, userId string) (companyID *int64, err error)
	AddUserToCompanyByEmail(ctx context.Context, companyID string, newUserRequest AddUser) error
	UpdateUserPermission(ctx context.Context, companyID string, userPermissionRequest UserPermissionRequest) error
	DeleteUser(ctx context.Context, companyID string, userPermissionRequest UserPermissionRequest) error
	SearchCompany(ctx context.Context, query string) (queryResult []Company, err error)
	GetCompanyUsers(ctx context.Context, companyID string) ([]CompanyUserResponse, error)
	GetUserCompanies(ctx context.Context, userID string) (companies []Company, err error)

	// Team
	CreateTeam(ctx context.Context, team Team) (err error)
	AddUserToTeam(ctx context.Context) error
}

type CompanyClient struct {
	DB *sql.DB
}

// NewCompanyClient creates a new company client
func NewCompanyClient(DB *sql.DB) *CompanyClient {
	return &CompanyClient{DB: DB}
}

const getUserCompaniesQuery = `
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

func (c *CompanyClient) GetUserCompanies(ctx context.Context, userID string) (companies []Company, err error) {
	rows, err := c.DB.QueryContext(ctx, getUserCompaniesQuery, userID)
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
	INSERT INTO USER_COMPANY(CompanyId,UserId,AccessId)
	VALUES ($1,$2,$3)
`

// AddUser adds a user by email address
func (c *CompanyClient) AddUserToCompanyByEmail(ctx context.Context, companyID string, newUserRequest AddUser) (err error) {

	user, err := utils.FindUserByEmail(ctx, c.DB, newUserRequest.Email)
	if err != nil {
		return
	}

	accessID, err := newUserRequest.Access.ToAccessID()
	if err != nil {
		return errors.WithStack(err)
	}
	res, err := c.DB.ExecContext(ctx, addUserToCompanyByEmailQuery, companyID, user.ID, accessID)
	if err != nil {
		return
	}

	nrows, err := res.RowsAffected()
	if err != nil || nrows == 0 {
		return errors.New("not able to add user")
	}
	return nil
}

const updateUserPermissionQuery = `
	UPDATE USER_COMPANY
	SET AccessId=$3
	WHERE CompanyId=$1 AND UserId=$2
`

func (c *CompanyClient) UpdateUserPermission(ctx context.Context, companyID string, userPermissionRequest UserPermissionRequest) (err error) {
	accessID, err := userPermissionRequest.Access.ToAccessID()
	if err != nil {
		return errors.WithStack(err)
	}

	res, err := c.DB.ExecContext(ctx, updateUserPermissionQuery, companyID, userPermissionRequest.UserID, accessID)
	if err != nil {
		return
	}

	nrows, err := res.RowsAffected()
	if err != nil || nrows == 0 {
		return errors.New("not able to add user")
	}
	return nil
}

const deleteUserQuery = `
	DELETE FROM USER_COMPANY
	WHERE CompanyId=$1 AND UserId=$2;
`

func (c *CompanyClient) DeleteUser(ctx context.Context, companyID string, user UserPermissionRequest) (err error) {
	res, err := c.DB.ExecContext(ctx, deleteUserQuery, companyID, user.UserID)
	if err != nil {
		return errors.WithStack(err)
	}
	nrows, err := res.RowsAffected()
	if err != nil || nrows == 0 {
		return errors.New("not able to delete user")
	}
	return
}

const getCompanyUsersQuery = `
	SELECT 
	uc.UserId
	,u.Username
	,ca.AccessLevel
	FROM USER_COMPANY uc
	LEFT JOIN(
		SELECT 
		ID
		,Username
		FROM Users
	) as u
	ON u.ID=uc.UserId
	LEFT JOIN(
		SELECT AccessLevel,
		AccessID
		FROM COMPANY_ACCESS
	) as ca
	ON ca.AccessID=uc.AccessID
	WHERE uc.CompanyId=$1
`

func (c *CompanyClient) GetCompanyUsers(ctx context.Context, companyID string) (responses []CompanyUserResponse, err error) {
	rows, err := c.DB.QueryContext(ctx, getCompanyUsersQuery, companyID)
	if err != nil {
		return responses, errors.WithStack(err)
	}
	defer rows.Close()

	for rows.Next() {
		var companyUser CompanyUserResponse
		if err = rows.Scan(&companyUser.UserID, &companyUser.Username, &companyUser.Access); err != nil {
			return responses, errors.WithStack(err)
		}
		responses = append(responses, companyUser)
	}

	if err = rows.Close(); err != nil {
		return responses, errors.WithStack(err)
	}
	if err = rows.Err(); err != nil {
		return responses, errors.WithStack(err)
	}
	return
}

const searchCompanyQuery = `
	SELECT ID
	,Name
	FROM COMPANY
	WHERE Name like $1
`

// SearchCompany receives a string query and returns a list of company results
func (c *CompanyClient) SearchCompany(ctx context.Context, query string) (queryResult []Company, err error) {
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
	INSERT INTO USER_COMPANY(CompanyId,UserId,AccessID)
	VALUES ($1,$2,1)
`

// CreateCompany creates a new company and adds the current user to it
func (c *CompanyClient) CreateCompany(ctx context.Context, companyName, userID string) (companyID *int64, err error) {
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

func (c *CompanyClient) CreateTeam(ctx context.Context, team Team) (err error) {

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
func (c *CompanyClient) AddUserToTeam(ctx context.Context) error {
	return nil
}
