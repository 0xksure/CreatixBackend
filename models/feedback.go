package models

import (
	"context"
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Feedback struct {
	UserID      uint
	Title       string `gorm:"type:varchar(264)"`
	Description string
}

type Clap struct {
	UserID     uint
	FeedbackID uint
	User       User `gorm:"foreignkey:ID;association_foreignkey:UserID"`
}

type Comment struct {
	UserID     uint
	FeedbackID uint
	Comment    string
	User       User `gorm:"foreignkey:ID;association_foreignkey:UserID"`
}

var createFeedback = `
	INSERT INTO FEEDBACK 
	VALUES ( $1, $2, $3 );
`

// CreateFeedback inserts the feedback into the database
func (f Feedback) CreateFeedback(ctx context.Context, db *sql.DB) (err error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	res, err := tx.Exec(createFeedback, f.UserID, f.Title, f.Description)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
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
	DELETE FROM FEEDBACK WHERE ID=$1;
`

// DeleteFeedback deletes feedback written by user
func (f Feedback) DeleteFeedback(ctx context.Context, db *sql.DB, id string) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	res, err := tx.Exec(deleteFeedback, id)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
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
	SET Title=$2,Description=$3
	WHERE ID=$1
`

// UpdateFeedback updates the database
func (f Feedback) UpdateFeedback(ctx context.Context, db *sql.DB, id string, feedback Feedback) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	res, err := tx.Exec(updateFeedback, id, feedback.Title, feedback.Description)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return err
		}
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

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("0 rows affected")
	}

	return nil
}

// ClapFeedback gives claps to feedback based on id, whomever can clap a feedback
func (f Feedback) ClapFeedback(db *gorm.DB, userEmail string, feedbackID string) (Feedback, error) {
	tx := db.Begin()
	db.Preload("Feedbacks")
	var clap Clap
	var clappingUser User
	var feedbackToBeClapped Feedback
	// Find user that wants to clap
	res := tx.Where("email = ?", userEmail).First(&clappingUser)
	if err := res.Error; err != nil {
		tx.Rollback()
		return feedbackToBeClapped, errors.Wrap(err, "not able to find user")
	}
	clap.User = clappingUser
	clap.UserID = clappingUser.ID

	// Find feedback to be clapped
	res = tx.Where("id = ?", feedbackID).First(&feedbackToBeClapped)
	if err := res.Error; err != nil {
		tx.Rollback()
		return feedbackToBeClapped, errors.Wrap(err, "not able to find feedback to clap")
	}
	clap.FeedbackID = feedbackToBeClapped.ID

	// Check if clap allready exists
	var clapped Clap
	var count int
	res = tx.Unscoped().Where("user_id = ? AND feedback_id = ?", clap.UserID, feedbackID).Find(&clapped).Count(&count)
	if count > 0 {
		// A clap already exists
		if clapped.DeletedAt == nil {
			// delete clap
			res = db.Delete(&clapped)
			if err := res.Error; err != nil {
				tx.Rollback()
				return feedbackToBeClapped, errors.Wrap(err, "not able to delete clap")
			}
			return feedbackToBeClapped, nil
		}
	}

	// create clap
	res = tx.Create(&clap)
	if err := res.Error; err != nil {
		tx.Rollback()
		return feedbackToBeClapped, errors.Wrap(err, "not able to create clap")
	}

	resAssociation := tx.Model(&feedbackToBeClapped).Association("Claps").Append(&clap)
	if err := resAssociation.Error; err != nil {
		tx.Rollback()
		return feedbackToBeClapped, errors.Wrap(err, "not able to append clap")
	}

	return feedbackToBeClapped, tx.Commit().Error
}

// GetUserFeedback returns the feedback created by the given user
func (f Feedback) GetUserFeedback(db *gorm.DB, userEmail string) ([]Feedback, error) {
	var feedbacks []Feedback
	var user User
	if err := db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return feedbacks, errors.Wrap(err, "cannot find user")
	}
	db.Where("user_id = ?", user.ID).Preload("Claps").Preload("Claps.User").
		Preload("Comments").Preload("Comments.User").
		Preload("User").Find(&feedbacks)

	return feedbacks, nil
}

// CommentFeedback writes a comment on the feedback
func (f Feedback) CommentFeedback(db *gorm.DB, userEmail string, feedbackID string, comment Comment) error {
	// Find user that wants to comment
	tx := db.Begin()
	var user User
	res := tx.Where("email = ?", userEmail).First(&user)
	if err := res.Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "not able to find user")
	}

	// FInd feedback to be
	var feedback Feedback
	res = tx.Where("id = ?", feedbackID).First(&feedback)
	if err := res.Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "not able to find feedback")
	}

	// Create comment
	comment.UserID = user.ID
	comment.FeedbackID = feedback.ID
	res = tx.Create(&comment)
	if err := res.Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "not able to create comment ")
	}

	// append to feedback
	resAssociation := tx.Model(&feedback).Association("Comments").Append(&comment)
	if err := resAssociation.Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "not able to append comment")
	}

	return tx.Commit().Error
}

func (f Feedback) UpdateComment(db *gorm.DB, commentID string, comment Comment) error {
	tx := db.Begin()
	res := tx.Model(&comment).Where("id = ?", commentID).Update("comment", comment.Comment)
	if err := res.Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "not able to updat comment ")
	}

	return tx.Commit().Error
}
