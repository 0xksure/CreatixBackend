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
)

type WebSocketRequest struct {
	Action     WebSocketAction `json:"action"`
	FeecbackID string          `json:"feedbackId"`
	Feedback   Feedback        `json:"feedback"`
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

type Clap struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	FeedbackID string `json:"feedbackId`
}

type Comment struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
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
	INSERT INTO FEEDBACK(UserID,Title,Description)
	VALUES ( $1, $2, $3 );
`

// CreateFeedback inserts the feedback into the database
func (f Feedback) CreateFeedback(ctx context.Context, db *sql.DB) (err error) {
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
	uid, err := strconv.Atoi(f.UserID)
	if err != nil {
		return
	}
	res, err := tx.Exec(createFeedback, uid, f.Title, f.Description)
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
func (f Feedback) DeleteFeedback(ctx context.Context, db *sql.DB, id string) error {
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

	currentTime := time.Now()
	res, err := tx.Exec(deleteFeedback, currentTime.Format(time.RFC3339), id)
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

var updateFeedback = `
	UPDATE FEEDBACK
	SET Title=$2,Description=$3,UpdatedAt=$4
	WHERE ID=$1
`

// UpdateFeedback updates the database
func (f Feedback) UpdateFeedback(ctx context.Context, db *sql.DB) error {
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

	currentTime := time.Now()
	res, err := tx.Exec(updateFeedback, f.ID, f.Title, f.Description, currentTime.Format(time.RFC3339))
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
func (f Feedback) ClapFeedback(ctx context.Context, db *sql.DB, userID string, feedbackID string) error {
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

func (fs *Feedbacks) GetUserClaps(ctx context.Context, db *sql.DB) error {
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

	for idx, f := range *fs {
		c, err := getClaps(ctx, tx, f.ID)
		if err != nil {
			return err
		}
		(*fs)[idx].Claps = c
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
func (f Feedback) GetUserFeedback(ctx context.Context, db *sql.DB, userID string) (Feedbacks, error) {
	var feedbacks Feedbacks
	uid, err := strconv.Atoi(userID)
	if err != nil {
		return feedbacks, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
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

const commentFeedback = `
	INSERT INTO COMMENTS(UserID,FeedbackID,Comment)
	SELECT $1,$2,$3
`

// CommentFeedback writes a comment on the feedback
func (c Comment) CommentFeedback(ctx context.Context, db *sql.DB, UserID string) (err error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return
	}

	uid, err := strconv.Atoi(UserID)
	if err != nil {
		return err
	}

	fid, err := strconv.Atoi(c.FeedbackID)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	res, err := tx.ExecContext(ctx, commentFeedback, uid, fid, c.Comment)
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
func (c Comment) UpdateComment(ctx context.Context, db *sql.DB) error {
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

	cid, err := strconv.Atoi(c.ID)
	if err != nil {
		return err
	}
	currentTime := time.Now()
	res, err := tx.ExecContext(ctx, updateComment, cid, c.Comment, currentTime.Format(time.RFC3339))
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

func (fs *Feedbacks) GetUserComments(ctx context.Context, db *sql.DB) error {
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

	for idx, f := range *fs {
		c, err := getComments(ctx, tx, f.ID)
		if err != nil {
			return err
		}
		(*fs)[idx].Comments = c
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
		if err := rows.Scan(&comment.ID, &comment.Comment, &comment.UserID, &comment.Person.Firstname, &comment.Person.Lastname); err != nil {
			return comments, err
		}

		comments = append(comments, comment)
	}

	return comments, nil
}
