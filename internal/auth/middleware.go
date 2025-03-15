package auth

import (
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

		// Try to parse as staff token
		staffClaims := &StaffClaims{}
		staffToken, err := jwt.ParseWithClaims(tokenString, staffClaims, func(token *jwt.Token) (interface{}, error) {
			return config.ConfigInstance.JWTSecret(), nil
		})

		if err == nil && staffToken.Valid {
			// It's a valid staff token
			c.Set("claims", staffClaims)
			c.Set("role", staffClaims.Role)
			c.Set("username", staffClaims.Username)
			c.Set("tokenType", "staff")
			c.Next()
			return
		}

		// Try to parse as customer token
		customerClaims := &CustomerClaims{}
		customerToken, err := jwt.ParseWithClaims(tokenString, customerClaims, func(token *jwt.Token) (interface{}, error) {
			return config.ConfigInstance.JWTSecret(), nil
		})

		if err == nil && customerToken.Valid {
			// It's a valid anonymous token
			c.Set("claims", customerClaims)
			c.Set("orderID", customerClaims.OrderID)
			c.Set("tokenType", "customer")
			c.Next()
			return
		}

		// If we get here, token is invalid
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
	}
}

// RequireRole creates middleware that checks if the user has a specific role
func RequireRole(role models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenType, exists := c.Get("tokenType")
		if !exists || tokenType != "staff" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Staff access required"})
			return
		}

		userRole, exists := c.Get("role")
		if !exists || userRole.(models.Role) != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}

// RequireOrderAccess creates middleware that checks if the person is authorized to modify the order based on id
func RequireOrderAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenType, exists := c.Get("tokenType")
		if !exists || tokenType != "customer" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Modification of this order is not allowed"})
			return
		}

		contextOrderID, exists := c.Get("orderID")

		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Modification of this order is not allowed"})
			return
		}

		pathOrderID := c.Param("orderID")

		if contextOrderID != pathOrderID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Modification of this order is not allowed"})
			return
		}

		c.Next()
	}
}
