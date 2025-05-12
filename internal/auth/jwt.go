package auth

import (
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

const (
	CustomerJwtExpirationTime = 4 * time.Hour
	StaffJwtExpirationTime    = 18 * time.Hour
)

// JWTType represents the type of JWT token. It can be either staff or customer.
// The type is used to determine the type of claims in the token.
// Staff JWT tokens hold information about authenticated staff members
// Customer JWT tokens are used to authorize review postings and order modifications by anonymous customers.
type JWTType string

const (
	StaffJWT    JWTType = "staff"
	CustomerJWT JWTType = "customer"
)

// StaffClaims represents the claims in a staff JWT token
type StaffClaims struct {
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}

// CustomerClaims represents the claims in an anonymous JWT token
type CustomerClaims struct {
	OrderID string `json:"order_id"`
	jwt.RegisteredClaims
}

// GenerateStaffJWT generates a JWT token for staff members
// and returns the encoded token, expiration time and error
func GenerateStaffJWT(username string, role models.Role) (string, time.Time, error) {
	expirationTime := time.Now().Add(StaffJwtExpirationTime)

	claims := StaffClaims{
		Role:             role,
		RegisteredClaims: newRegisteredClaims(username, expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = StaffJWT

	tokenStr, err := token.SignedString([]byte(config.ConfigInstance.JWTSecret()))
	if err != nil {
		return "", time.Time{}, errors.NewInternalServerErr("Failed to generate JWT token", err)
	}

	return tokenStr, expirationTime, err

}

// GenerateCustomerJWT generates a JWT token for anonymous customers
// and returns the encoded token, expiration time and error
func GenerateCustomerJWT(orderID uint) (string, time.Time, error) {
	expirationTime := time.Now().Add(CustomerJwtExpirationTime)
	claims := CustomerClaims{
		OrderID:          strconv.FormatUint(uint64(orderID), 10),
		RegisteredClaims: newRegisteredClaims("anonymous", expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["typ"] = CustomerJWT

	encodedToken, err := token.SignedString([]byte(config.ConfigInstance.JWTSecret()))
	if err != nil {
		return "", time.Time{}, errors.NewInternalServerErr("Failed to generate a customer jwt", err)
	}
	return encodedToken, expirationTime, err
}

// getSecret returns the secret used to sign the JWT tokens.
func getSecret(_ *jwt.Token) (interface{}, error) {
	return config.ConfigInstance.JWTSecret(), nil
}

func newRegisteredClaims(subject string, expirationTime time.Time) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "mymeals-api",
		Subject:   subject,
	}
}
