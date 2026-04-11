package auth

import (
	"douyin-backend/app/global/variable"
	"douyin-backend/app/http/middleware/my_jwt"
	"github.com/gin-gonic/gin"
)

func GetUidFromToken(ctx *gin.Context) (uid int64) {
	tokenKey := variable.ConfigYml.GetString("Token.BindContextKeyName")
	userToken, exists := ctx.Get(tokenKey)
	if exists {
		uid = userToken.(my_jwt.CustomClaims).UID
	} else {
		uid = variable.ConfigYml.GetInt64("Token.JwtDefaultUid")
		variable.ZapLog.Error(ctx.ClientIP() + " userToken.UID not exists!")
	}
	return
}
