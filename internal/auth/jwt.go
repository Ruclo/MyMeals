package auth

import (
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

const (
	CustomerJwtExpirationTime = 4 * time.Hour
	StaffJwtExpirationTime    = 18 * time.Hour
)

type JWTType string

const (
	StaffJWT    JWTType = "staff"
	CustomerJWT JWTType = "customer"
)

// StaffClaims represents the claims in a staff JWT token
type StaffClaims struct {
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

// CustomerClaims represents the claims in an anonymous JWT token
type CustomerClaims struct {
	OrderID string `json:"order_id"`
	jwt.RegisteredClaims
}

// GenerateStaffToken generates a JWT token for staff members
func GenerateStaffToken(username string, role models.Role) (string, error) {
	claims := StaffClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(StaffJwtExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mymeals-api",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = StaffJWT
	return token.SignedString([]byte(config.ConfigInstance.JWTSecret()))
}

// GenerateCustomerToken generates a JWT token for customers
func GenerateCustomerToken(orderID uint) (string, error) {
	claims := CustomerClaims{
		OrderID: strconv.FormatUint(uint64(orderID), 10),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(CustomerJwtExpirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mymeals-api",
			Subject:   "anonymous",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = CustomerJWT
	return token.SignedString(config.ConfigInstance.JWTSecret())
}
