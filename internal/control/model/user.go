package model

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Creditionals struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type SetRole struct {
	UserId int `json:"user_id"`
	RoleId int `json:"role_id"`
}

type UserContext struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Role  string `json:"role_name"`
}

func NewToken(secret []byte, login string, role string, now time.Time) (*Token, error) {
	claims := jwt.MapClaims{
		"login": login,
		"role":  role,
		"exp":   now.Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, err
	}
	refreshClaims := jwt.MapClaims{
		"login":   login,
		"role":    role,
		"refresh": true,
		"exp":     now.Add(time.Hour * 24 * 365).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refTokenString, err := refreshToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return &Token{
		Token:        tokenString,
		RefreshToken: refTokenString,
	}, nil
}
