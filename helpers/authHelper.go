package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUSerType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	if userType == role {
		return err
	}
	if userType != role {
		err = errors.New("unauthorised to access this resource")
		return err
	}
	return err
}

func MatchUserTypeToid(c *gin.Context, userId string) (err error) {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil

	if userType == "USER" && uid != userId {
		err = errors.New("unauthorised to access this resource")
		return err
	}
	err = CheckUSerType(c, userType)
	return err
}
