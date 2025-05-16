package apperrors

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
)

// ErrorHandler handles app errors and returns the appropriate response.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last().Err
			log.Printf("Error: %v", err.Error())

			// Check if it's an AppError
			// All apperrors should be wrapped in AppError
			var appErr *AppError
			if !errors.As(err, &appErr) {
				log.Printf("Unhandled error: %v", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}

			// Check if it's a validation error
			if IsValidationErr(appErr) {
				var validationErrors validator.ValidationErrors
				errors.As(appErr.Err, &validationErrors)
				var invalidFields []string

				for _, err := range validationErrors {
					invalidFields = append(invalidFields, err.Field())
				}

				c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message, "invalidFields": invalidFields})
				return

			}

			c.JSON(appErr.StatusCode, appErr.Message)
			return

		}
	}
}
