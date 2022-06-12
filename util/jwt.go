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

// GetUserIDFormToken ParseToken 验证用户token
// id int64: 用户id 如果没有解析出，默认为-1
// err error: 错误
func GetUserIDFormToken(token string) (int64, error) {
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

// ValidToken 校验token是否过期
// bool: 是否过期 default: true 过期
// error: 解析是否成功 default: nil
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
