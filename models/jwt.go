package models

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type UserClaims struct {
	Id         int       `json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	CreateTime time.Time `json:"create_time"`
	Exp        int64     `json:"exp"`
	Iat        int64     `json:"iat"`
	jwt.StandardClaims
}
