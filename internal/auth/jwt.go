package auth

import (
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const (
	CUSTOMER_JWT_EXPIRATION_TIME = 4 * time.Hour
	STAFF_JWT_EXPIRATION_TIME    = 18 * time.Hour
)

// StaffClaims represents the claims in a staff JWT token
type StaffClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// CustomerClaims represents the claims in an anonymous JWT token
type CustomerClaims struct {
	OrderID string `json:"order_id"`
	jwt.RegisteredClaims
}

// GenerateStaffToken generates a JWT token for staff members
func GenerateStaffToken(username, role string) (string, error) {
	claims := StaffClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(STAFF_JWT_EXPIRATION_TIME)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mymeals-api",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.ConfigInstance.JWTSecret()))
}

// GenerateCustomerToken generates a JWT token for customers
func GenerateCustomerToken(orderID string) (string, error) {
	claims := CustomerClaims{
		OrderID: orderID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(CUSTOMER_JWT_EXPIRATION_TIME)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mymeals-api",
			Subject:   "anonymous",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.ConfigInstance.JWTSecret()))
}
