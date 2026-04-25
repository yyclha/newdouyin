package my_jwt

import (
	"github.com/dgrijalva/jwt-go"
)

// CustomClaims 定义系统使用的 JWT 自定义声明结构。
type CustomClaims struct {
	UID      int64  `json:"uid"`
	NickName string `json:"nickname"`
	Phone    string `json:"phone"`
	jwt.StandardClaims
}
