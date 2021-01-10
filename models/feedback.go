package models

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type WebSocketAction int

const (
	NewFeedback     WebSocketAction = 1
	ClapFeedback    WebSocketAction = 2
	CommentFeedback WebSocketAction = 3
	UpdateComment   WebSocketAction = 4
)

type FeedbackClienter interface {
	CreateFeedback(ctx context.Context, UserID, companyID string, feedback FeedbackRequest) (err error)
	DeleteFeedback(ctx context.Context, feedbackID string) error
	UpdateFeedback(ctx context.Context, feedbackID string, feedback FeedbackRequest) error

	IsUserOwnerOfFeedback(ctx context.Context, feedbackID, userID string) (isOwner bool, err error)

	GetUserFeedback(ctx context.Context, userID string) (Feedbacks, error)
	GetUserFeedbackwData(ctx context.Context, userID string) (feedbacks Feedbacks, err error)

	GetCompanyFeedbacks(ctx context.Context, companyID string) (feedbacks []Feedback, err error)
	GetCompanyFeedbackswData(ctx context.Context, companyID string) (feedbacks Feedbacks, err error)

	ClapFeedback(ctx context.Context, userID string, feedbackID string) error
	GetUserClaps(ctx context.Context) error

	CommentFeedback(ctx context.Context, comment, userID, feedbackID string) (err error)
	UpdateComment(ctx context.Context, commentID, comment string)
	GetUserComments(ctx context.Context, feedbacks []Feedback) error
}

type FeedbackClient struct {
	db *sql.DB
}

func NewFeedbackClient(db *sql.DB) *FeedbackClient {
	return &FeedbackClient{db}
}

type WebSocketRequest struct {
	Action     WebSocketAction `json:"action"`
	FeecbackID string          `json:"feedbackId"`
	Feedback   FeedbackRequest `json:"feedback"`
	Comment    Comment         `json:"comment"`
}

type Feedback struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Person      Person    `json:"person"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Comments    []Comment `json:"comments"`
	Claps       []Clap    `json:"claps"`
	UpdatedAt   *string   `json:"updatedAt`
}

type FeedbackRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Clap struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	FeedbackID string `json:"feedbackId`
}
type CommentRequest struct {
	Comment string `json:"comment"`
}

type Comment struct {
	ID         string `json:"id"`
	FeedbackID string `json:"feedbackId"`
	Person     Person `json:"person"`
	Comment    string `json:"comment"`
}

type Person struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

var createFeedback = `
	INSERT INTO FEEDBACK(UserID,CompanyID,Title,Description)
	VALUES ( $1, $2, $3, $4 );
`

// CreateFeedback inserts the feedback into the database
func (c *FeedbackClient) CreateFeedback(ctx context.Context, UserID, companyID string, feedback FeedbackRequest) (err error) {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	res, err := tx.Exec(createFeedback, UserID, companyID, feedback.Title, feedback.Description)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("0 rows affected")
	}

	return nil
}

var deleteFeedback = `
	UPDATE FEEDBACK SET DeletedAt=$1 WHERE ID=$2
`

// DeleteFeedback deletes feedback written by user
func (c *FeedbackClient) DeleteFeedback(ctx context.Context, feedbackID string) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	currentTime := time.Now()
	res, err := tx.Exec(deleteFeedback, currentTime.Format(time.RFC3339), feedbackID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("0 rows affected")
	}

	return nil
}

var isUserOwnerOfFeedbackQuery = `
	SELECT UserID
	FROM FEEDBACK 
	WHERE UserID=$1 AND Id=$2
`

func (c *FeedbackClient) IsUserOwnerOfFeedback(ctx context.Context, feedbackID, userID string) (isOwner bool, err error) {
	var exists bool
	err = c.db.QueryRowContext(ctx, isUserOwnerOfFeedbackQuery, userID, feedbackID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("user is not owner of feedback")
		}
		return
	}
	return
}

var updateFeedback = `
	UPDATE FEEDBACK
	SET Title=$2,Description=$3,UpdatedAt=$4
	WHERE ID=$1
`

// UpdateFeedback updates the database
func (c *FeedbackClient) UpdateFeedback(ctx context.Context, feedbackID string, feedback FeedbackRequest) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	currentTime := time.Now()
	res, err := tx.Exec(updateFeedback, feedbackID, feedback.Title, feedback.Description, currentTime.Format(time.RFC3339))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("0 rows affected")
	}

	return nil
}

var clapFeedback = ` 
	WITH upsert AS (
		UPDATE CLAPS SET deletedat=NOW()
		WHERE DeletedAt IS NULL AND (userid=$1 AND feedbackid=$2)
		RETURNING *
   )
   INSERT INTO CLAPS (UserID,FeedbackID)
   SELECT $1, $2 
   WHERE NOT EXISTS (SELECT * FROM upsert);
   
`

// ClapFeedback gives claps to feedback based on id, whomever can clap a feedback
func (c *FeedbackClient) ClapFeedback(ctx context.Context, userID string, feedbackID string) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	uid, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}

	fid, err := strconv.Atoi(feedbackID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, clapFeedback, uid, fid)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

type Feedbacks []Feedback

func (c *FeedbackClient) GetUserClaps(ctx context.Context, feedbacks []Feedback) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()
	for idx, f := range feedbacks {
		c, err := getClaps(ctx, tx, f.ID)
		if err != nil {
			return err
		}
		(feedbacks)[idx].Claps = c
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

const getClapsQuery = `
	SELECT 
	ID
	,UserID
	,FeedbackId
	FROM CLAPS
	WHERE FeedbackID=$1 AND DeletedAt IS NULL
`

func getClaps(ctx context.Context, tx *sql.Tx, feedbackID string) ([]Clap, error) {
	claps := make([]Clap, 0)
	fid, err := strconv.Atoi(feedbackID)
	if err != nil {
		return claps, err
	}

	rows, err := tx.QueryContext(ctx, getClapsQuery, fid)
	if err != nil {
		return claps, err
	}
	defer rows.Close()
	for rows.Next() {
		var clap Clap
		if err := rows.Scan(&clap.ID, &clap.UserID, &clap.FeedbackID); err != nil {
			return claps, err
		}

		claps = append(claps, clap)
	}

	return claps, nil
}

const getUserFeedback = `
	SELECT
	f.ID
	,f.UserID
	,u.Firstname
	,u.Lastname
	,f.Title
	,f.Description
	,f.UpdatedAt
	FROM FEEDBACK as f
	LEFT JOIN (
		SELECT
		ID 
		,Firstname
		,Lastname
		FROM USERS
		WHERE ID=$1
	) as u 
	ON u.ID=f.UserID
	WHERE UserID=$1 AND DeletedAt IS NULL
`

// GetUserFeedback returns the feedback created by the given user
func (c *FeedbackClient) GetUserFeedback(ctx context.Context, userID string) (Feedbacks, error) {
	var feedbacks Feedbacks
	uid, err := strconv.Atoi(userID)
	if err != nil {
		return feedbacks, err
	}
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return feedbacks, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	rows, err := tx.QueryContext(ctx, getUserFeedback, uid)
	if err != nil {
		return feedbacks, err
	}

	defer rows.Close()
	for rows.Next() {
		var feedback Feedback
		err = rows.Scan(&feedback.ID,
			&feedback.UserID,
			&feedback.Person.Firstname,
			&feedback.Person.Lastname,
			&feedback.Title,
			&feedback.Description,
			&feedback.UpdatedAt,
		)
		if err != nil {
			return feedbacks, err
		}
		feedbacks = append(feedbacks, feedback)
	}

	err = rows.Close()
	if err != nil {
		return feedbacks, err
	}

	if err := rows.Err(); err != nil {
		return feedbacks, err
	}
	err = tx.Commit()
	if err != nil {
		return feedbacks, err
	}

	return feedbacks, nil
}

func (c *FeedbackClient) GetCompanyFeedbackswData(ctx context.Context, companyID string) (feedbacks Feedbacks, err error) {
	feedbacks, err = c.GetCompanyFeedbacks(ctx, companyID)
	if err != nil {
		return
	}

	err = c.GetUserComments(ctx, feedbacks)
	if err != nil {
		return
	}

	err = c.GetUserClaps(ctx, feedbacks)
	if err != nil {
		return
	}
	return feedbacks, nil
}

func (c *FeedbackClient) GetUserFeedbackwData(ctx context.Context, userID string) (feedbacks Feedbacks, err error) {
	feedbacks, err = c.GetUserFeedback(ctx, userID)
	if err != nil {
		return
	}

	err = c.GetUserComments(ctx, feedbacks)
	if err != nil {
		return
	}

	err = c.GetUserClaps(ctx, feedbacks)
	if err != nil {
		return
	}
	return feedbacks, nil
}

const getFeedbackQuery = `
	SELECT
	f.ID
	,f.UserID
	,u.Firstname
	,u.Lastname
	,f.Title
	,f.Description
	,f.UpdatedAt
	FROM FEEDBACK as f
	LEFT JOIN (
		SELECT
		ID 
		,Firstname
		,Lastname
		FROM USERS
		WHERE ID=$1
	) as u 
	ON u.ID=f.UserID
	WHERE CompanyID=$1 AND DeletedAt IS NULL
`

// GetCompanyFeedbacks get the feedbacks for the given companyID
func (c *FeedbackClient) GetCompanyFeedbacks(ctx context.Context, companyID string) (feedbacks []Feedback, err error) {
	rows, err := c.db.QueryContext(ctx, getFeedbackQuery, companyID)
	if err != nil {
		return
	}
	defer rows.Close()

	var feedback Feedback
	for rows.Next() {
		err = rows.Scan(&feedback.ID,
			&feedback.UserID,
			&feedback.Person.Firstname,
			&feedback.Person.Lastname,
			&feedback.Title,
			&feedback.Description,
			&feedback.UpdatedAt,
		)
		if err != nil {
			return
		}
		feedbacks = append(feedbacks, feedback)
	}

	err = rows.Close()
	if err != nil {
		return
	}

	if err = rows.Err(); err != nil {
		return
	}

	return

}

const commentFeedback = `
	INSERT INTO COMMENTS(UserID,FeedbackID,Comment)
	SELECT $1,$2,$3
`

// CommentFeedback writes a comment on the feedback
func (c *FeedbackClient) CommentFeedback(ctx context.Context, comment, userID, feedbackID string) (err error) {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	uid, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}

	fid, err := strconv.Atoi(feedbackID)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	res, err := tx.ExecContext(ctx, commentFeedback, uid, fid, comment)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	nrows, err := res.RowsAffected()
	if err != nil {
		return
	}
	if nrows == 0 {
		return errors.New("0 rows affected")
	}

	return
}

const updateComment = `
UPDATE COMMENTS SET Comment=$2,UpdatedAt=$3 WHERE ID=$1
`

// UpdateComment updates a comment based on the comment ID
func (c *FeedbackClient) UpdateComment(ctx context.Context, commentID, comment string) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	cid, err := strconv.Atoi(commentID)
	if err != nil {
		return err
	}
	currentTime := time.Now()
	res, err := tx.ExecContext(ctx, updateComment, cid, comment, currentTime.Format(time.RFC3339))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.New("0 rows affected")
	}

	return nil
}

func (c *FeedbackClient) GetUserComments(ctx context.Context, feedbacks []Feedback) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	for idx, f := range feedbacks {
		c, err := getComments(ctx, tx, f.ID)
		if err != nil {
			return err
		}
		(feedbacks)[idx].Comments = c
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

const getCommentsQuery = `
SELECT 
c.ID
,c.comment
,c.FeedbackId 
,c.userid
,u.firstname
,u.lastname
FROM comments as c
LEFT JOIN 
(SELECT id,firstname,lastname FROM users ) as u on u.id = c.userid 
WHERE feedbackid=$1
`

func getComments(ctx context.Context, tx *sql.Tx, feedbackID string) ([]Comment, error) {
	comments := make([]Comment, 0)
	fid, err := strconv.Atoi(feedbackID)
	if err != nil {
		return comments, err
	}

	rows, err := tx.QueryContext(ctx, getCommentsQuery, fid)
	if err != nil {
		return comments, err
	}
	defer rows.Close()
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.Comment, &comment.FeedbackID, &comment.Person.ID, &comment.Person.Firstname, &comment.Person.Lastname); err != nil {
			return comments, err
		}

		comments = append(comments, comment)
	}

	return comments, nil
}
