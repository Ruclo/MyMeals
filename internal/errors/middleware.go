package errors

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			log.Printf("Error: %v", err.Error())

			var appErr *AppError
			if errors.As(err, &appErr) {
				c.JSON(appErr.StatusCode, appErr.Message)
			}
		}
	}
}
