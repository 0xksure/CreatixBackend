package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Feedback struct {
	gorm.Model
	UserID      uint
	Title       string `gorm:"type:varchar(264)"`
	Description string
	Claps       []Clap    `gorm:"foreignkey:feedback_id"`
	Comments    []Comment `gorm:"foreignkey:feedback_id"`
	User        User      `gorm:"foreignkey:ID;association_foreignkey:UserID;auto_preload"`
}

type Clap struct {
	gorm.Model
	UserID     uint
	FeedbackID uint
	User       User `gorm:"foreignkey:ID;association_foreignkey:UserID"`
}

type Comment struct {
	gorm.Model
	UserID     uint
	FeedbackID uint
	Comment    string
	User       User `gorm:"foreignkey:ID;association_foreignkey:UserID"`
}

// CreateFeedback inserts the feedback into the database
func (f *Feedback) CreateFeedback(db *gorm.DB, user *User) (createdFeedback Feedback, err error) {
	// create feedback
	if err = db.First(&user).Error; err != nil {
		return createdFeedback, errors.Wrap(err, "not able to find user")
	}
	f.User = *user
	f.UserID = user.ID
	if err = db.Create(&f).Error; err != nil {
		return createdFeedback, errors.Wrap(err, "not able to create feedback")
	}
	return *f, nil
}

// DeleteFeedback deletes feedback written by user
func (f *Feedback) DeleteFeedback(db *gorm.DB, user *User, id string) error {
	db.Preload("Feedbacks")
	// Find feedback with given if for user
	var feedbackToBeDeleted Feedback

	res := db.Model(f).First(&feedbackToBeDeleted, id)
	if err := res.Error; err != nil {
		return err
	}

	// Delete feedback from feedback
	res = db.Delete(&feedbackToBeDeleted)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// UpdateFeedback updates the database
func (f *Feedback) UpdateFeedback(db *gorm.DB, user *User, id string) (Feedback, error) {
	db.Preload("Feedbacks")
	var feedbackToBeUpdated Feedback
	res := db.Model(f).First(&feedbackToBeUpdated, id)
	if err := res.Error; err != nil {
		return feedbackToBeUpdated, errors.Wrap(err, "not able to find feedback")
	}
	// Update feedback in User
	feedbackToBeUpdated.Title = f.Title
	feedbackToBeUpdated.Description = f.Description

	// Update feedback in feedback
	fmt.Println(feedbackToBeUpdated)
	res = db.Save(&feedbackToBeUpdated)
	if res.Error != nil {
		return feedbackToBeUpdated, errors.Wrap(res.Error, "not able to update feedback in feedback")
	}
	return feedbackToBeUpdated, nil
}

// ClapFeedback gives claps to feedback based on id, whomever can clap a feedback
func (f *Feedback) ClapFeedback(db *gorm.DB, userEmail string, feedbackID string) (Feedback, error) {
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
func (f *Feedback) GetUserFeedback(db *gorm.DB, userEmail string) ([]Feedback, error) {
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
func (f *Feedback) CommentFeedback(db *gorm.DB, userEmail string, feedbackID string, comment Comment) error {
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

func (f *Feedback) UpdateComment(db *gorm.DB, commentID string, comment Comment) error {
	tx := db.Begin()
	res := tx.Model(&comment).Where("id = ?", commentID).Update("comment", comment.Comment)
	if err := res.Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "not able to updat comment ")
	}

	return tx.Commit().Error
}
