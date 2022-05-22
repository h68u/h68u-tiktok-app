package util

import (
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserId int64 `json:"userId"`
	jwt.StandardClaims
}

//GenerateToken 签发用户Token
func CreateAccessToken(userId int64) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(2 * 60 * time.Minute)
	claims := Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "tiktok-app",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func CreateRefreshToken(userId int64) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(30 * 24 * 60 * time.Minute)
	claims := Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "tiktok-app",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

//ParseToken 验证用户token
func GetUsernameFormToken(token string) (int64, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims.UserId, nil
		}
	}
	return -1, err
}

//校验token是否过期
func ValidToken(token string) (bool, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			expiresTime := claims.ExpiresAt
			now := time.Now().Unix()
			if now > expiresTime {
				//token过期了
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return true, err
}
