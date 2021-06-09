package webhandlers

import (
	"iot_meter/models"


	"github.com/gin-gonic/gin"
	"github.com/appleboy/gin-jwt/v2"
)

type loginForm struct {
	Email    string `form:"email" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func LoginHandler(c *gin.Context) (interface{}, error) {
	var loginVals loginForm
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	
		email := loginVals.Email
		password := loginVals.Password
		
		// TEMPORARY CODE TO GENERATE JWT
		if (email == "admin" && password == "admin") {
			return &models.Account{
				ID  : "MASTER",
				Email : "admin@goteleport.com",
				PlanType : "S",
				TotalUsers : 0,
			}, nil
		}

	return nil, jwt.ErrFailedAuthentication
}