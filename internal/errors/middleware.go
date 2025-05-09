package errors

import (
	stdErrors "errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			log.Printf("Error: %v", err.Error())

			var appErr *AppError
			if !stdErrors.As(err, &appErr) {
				log.Printf("Unhandled error: %v", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}

			var validationErrors validator.ValidationErrors
			if stdErrors.As(appErr.Err, &validationErrors) {
				var invalidFields []string

				for _, err := range validationErrors {
					invalidFields = append(invalidFields, err.Field())
				}

				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "invalidFields": invalidFields})
				return
			}

			c.JSON(appErr.StatusCode, appErr.Message)
			return

		}
	}
}
