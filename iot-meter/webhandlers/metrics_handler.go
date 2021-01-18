package webhandlers

import (
	"log"
	"time"
	"errors"

	"iot_meter/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type metricsPayload struct {
	AccountID   string `form:"account_id" json:"account_id" binding:"required"`
	UserID 		string `form:"user_id" json:"user_id" binding:"required"`
	Timestamp	time.Time `form:"timestamp" json:"timestamp" binding:"required"`
}

func Metrics(c *gin.Context) {
	
	var payload metricsPayload

	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "Missing required parameters " + err.Error(),
		})
	}

	var account models.Account
	var user models.User

	log.Println("Received metrics hit")

	db := c.Keys["DB"].(*gorm.DB)
	tx := db.FirstOrCreate(&account, models.Account{ID: payload.AccountID})
	if tx.Error != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "Fatal error creating new account " + tx.Error.Error(),
		})
	}

	tx = db.Where(&models.User{ID: payload.UserID, IotAccount: account}).First(&user)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "Fatal error creating new user " + tx.Error.Error(),
		})
	}

	// Ir we got a not found, we have to insert, first check wether we can
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		

		// Hopefully this will sync access to the db :)
		err := db.Transaction(func(tx *gorm.DB) error {
			
			// Reselect for update, so other transactions wait until commit
			res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where(models.Account{ID: payload.AccountID}).
					First(&account)

			if res.Error != nil {
				return res.Error
			}

			if !account.CanAcceptMoreUsers() {
				c.JSON(403, gin.H{
					"code": 403,
					"message": "Account plan limit reached",
				})
				return nil
			}
			
			res = tx.Create(&models.User{ID: payload.UserID, IotAccountID: account.ID, UpdatedAt: payload.Timestamp})
			if res.Error != nil {
				return res.Error
			}

			account.IncreaseUserCount()
			res = tx.Save(account)
			if res.Error != nil {
				return res.Error
			}

			// nil => commit
			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{
				"code":    500,
				"message": "Fatal error updating databass " + err.Error(),
			})
		}
		
	} else { // Otherwise just update TS, no need to make this locking

		user.UpdatedAt = payload.Timestamp
		tx := db.Save(user)

		if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.JSON(500, gin.H{
				"code":    500,
				"message": "Fatal updating user timestamp " + tx.Error.Error(),
			})
		}
	}

	// All ok, we can go ahead, create and 

	c.JSON(200, gin.H{
		"code": 200,
		"message": "Data stored",
	})
}
