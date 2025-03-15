package auth

import (
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"time"
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(18 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mymeals-api",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.ConfigInstance.JWTSecret()))
}

// GenerateCustomerToken generates a JWT token for anonymous users
func GenerateCustomerToken(orderID string) (string, error) {
	claims := CustomerClaims{
		OrderID: orderID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(18 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mymeals-api",
			Subject:   "anonymous",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.ConfigInstance.JWTSecret()))
}

// ValidateStaffToken validates a staff JWT token and returns its claims
func ValidateStaffToken(tokenString string) (*StaffClaims, error) {
	claims := &StaffClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.ConfigInstance.JWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// ValidateAnonymousToken validates an anonymous JWT token and returns its claims
func ValidateAnonymousToken(tokenString string) (*CustomerClaims, error) {
	claims := &CustomerClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.ConfigInstance.JWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
