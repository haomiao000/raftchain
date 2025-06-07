package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var MySecret = []byte("this is a very complex secret")

func GetKey(_ *jwt.Token) (interface{}, error) {
	return MySecret, nil
}

type MyClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// 生成token
func GenToken(userID int64) (string, error) { 
	claims := MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "jwt",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(MySecret)
}