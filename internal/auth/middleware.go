package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/models"
)

// AuthMiddleware creates Gin middleware for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return config.ConfigInstance.JWTSecret(), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		tokenType, ok := token.Header["typ"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
			return
		}

		switch JWTType(tokenType) {
		case StaffJWT:
			staffClaims := &StaffClaims{}
			token, err := jwt.ParseWithClaims(tokenString, staffClaims, func(token *jwt.Token) (interface{}, error) {
				return config.ConfigInstance.JWTSecret(), nil
			})

			if err != nil || !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				return
			}
			c.Set("role", staffClaims.Role)
			c.Set("username", staffClaims.Username)
			c.Set("tokenType", StaffJWT)
			c.Next()

		case CustomerJWT:
			customerClaims := &CustomerClaims{}
			token, err := jwt.ParseWithClaims(tokenString, customerClaims, func(token *jwt.Token) (interface{}, error) {
				return config.ConfigInstance.JWTSecret(), nil
			})
			if err != nil || !token.Valid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				return
			}
			c.Set("orderID", customerClaims.OrderID)
			c.Set("tokenType", CustomerJWT)
			c.Next()
		}
	}
}

// RequireAnyRole creates middleware that checks if the user has any of the speciefied roles
func RequireAnyRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenType, exists := c.Get("tokenType")
		fmt.Println(tokenType)
		if !exists || tokenType.(JWTType) != StaffJWT {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Staff access required"})
			return
		}

		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		for _, role := range roles {

			if userRole.(models.Role) == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}
}

// RequireOrderAccess creates middleware that checks if the person is authorized to modify the order based on id
func RequireOrderAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenType, exists := c.Get("tokenType")
		if !exists || tokenType != CustomerJWT {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Modification of this order is not allowed1"})
			return
		}

		contextOrderID, exists := c.Get("orderID")

		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Modification of this order is not allowed2"})
			return
		}

		pathOrderID := c.Param("orderID")

		if contextOrderID != pathOrderID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Modification of this order is not allowed3"})
			return
		}

		c.Next()
	}
}
