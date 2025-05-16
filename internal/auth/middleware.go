package auth

import (
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware makes sure token cookie is present and valid.
// Parses the token and sets the appropriate claims on the context.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.Error(apperrors.NewUnauthorizedErr("Missing auth cookie", err))
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, getSecret)

		if err != nil || !token.Valid {
			c.Error(apperrors.NewUnauthorizedErr("Invalid or expired token", err))
			c.Abort()
			return
		}

		tokenType, ok := token.Header["typ"].(string)
		if !ok {
			c.Error(apperrors.NewUnauthorizedErr("Invalid token type", nil))
			c.Abort()
			return
		}

		switch JWTType(tokenType) {
		case StaffJWT:
			staffClaims := &StaffClaims{}
			token, err = jwt.ParseWithClaims(tokenString, staffClaims, getSecret)

			if err != nil || !token.Valid {
				c.Error(apperrors.NewUnauthorizedErr("Invalid or expired token", err))
				c.Abort()
				return
			}

			c.Set("role", staffClaims.Role)
			c.Set("username", staffClaims.Subject)
			c.Set("tokenType", StaffJWT)
			c.Next()

		case CustomerJWT:
			customerClaims := &CustomerClaims{}
			token, err = jwt.ParseWithClaims(tokenString, customerClaims, getSecret)
			if err != nil || !token.Valid {
				c.Error(apperrors.NewUnauthorizedErr("Invalid or expired token", err))
				c.Abort()
				return
			}

			c.Set("orderID", customerClaims.OrderID)
			c.Set("tokenType", CustomerJWT)
			c.Next()
		}
	}
}

// RequireAnyRole middleware checks if the authenticated user has any of the specified roles.
func RequireAnyRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenType, exists := c.Get("tokenType")

		if !exists || tokenType.(JWTType) != StaffJWT {
			c.Error(apperrors.NewForbiddenErr("Staff access required", nil))
			c.Abort()
			return
		}

		userRole, exists := c.Get("role")
		if !exists {
			c.Error(apperrors.NewInternalServerErr("Missing role in context", nil))
			c.Abort()
			return
		}

		for _, role := range roles {
			if userRole.(models.Role) == role {
				c.Next()
				return
			}
		}

		c.Error(apperrors.NewForbiddenErr("Insufficient permissions", nil))
		c.Abort()
	}
}

// RequireOrderAccess creates middleware that checks if the person is authorized to modify the order based on id.
func RequireOrderAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenType, exists := c.Get("tokenType")
		if !exists || tokenType != CustomerJWT {
			c.Error(apperrors.NewForbiddenErr("Only the creator of this order can modify it", nil))
			c.Abort()
			return
		}

		contextOrderID, exists := c.Get("orderID")

		if !exists {
			c.Error(apperrors.NewInternalServerErr("Missing order id in context", nil))
			c.Abort()
			return
		}

		pathOrderID := c.Param("orderID")

		if contextOrderID != pathOrderID {
			c.Error(apperrors.NewForbiddenErr("Only the creator of this order can modify it", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
