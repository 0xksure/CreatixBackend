package models

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Feedback struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Clap struct {
	UserID     uint
	FeedbackID uint
	User       User
}

type Comment struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	FeedbackID string `json:"feedbackId"`
	Comment    string `json:"comment"`
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
	INSERT INTO CLAPS(UserID,FeedbackID)
	SELECT $1,$2
	WHERE NOT EXISTS (
		SELECT ID FROM CLAPS WHERE DeletedAt IS NOT NULL OR (userid=$1 AND feedbackid=$2)
	)
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

	res, err := tx.ExecContext(ctx, clapFeedback, uid, fid)
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

const getUserFeedback = `
	SELECT
	ID
	,UserID
	,Title
	,Description
	FROM FEEDBACK
	WHERE UserID=$1 AND DeletedAt IS NULL
`

// GetUserFeedback returns the feedback created by the given user
func (f Feedback) GetUserFeedback(ctx context.Context, db *sql.DB, userID string) ([]Feedback, error) {
	var feedbacks []Feedback
	uid, err := strconv.Atoi(userID)
	if err != nil {
		return feedbacks, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return feedbacks, err
	}
	rows, err := tx.QueryContext(ctx, getUserFeedback, uid)
	if err != nil {
		return feedbacks, err
	}

	defer rows.Close()
	for rows.Next() {
		var feedback Feedback
		err = rows.Scan(&feedback.ID, &feedback.UserID, &feedback.Title, &feedback.Description)
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
	return feedbacks, nil
}

const commentFeedback = `
	INSERT INTO COMMENTS(UserID,FeedbackID,Comment)
	VALUES ($1,$2,$3)
	WHERE NOT EXIST (
		SELECT UserID from COMMENTS WHERE UserID=$1
	)
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

func (c Comment) UpdateComment(ctx context.Context, db *sql.DB, commentID string, comment string) error {
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
	res, err := tx.ExecContext(ctx, commentID, comment, currentTime.Format(time.RFC3339))
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
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
