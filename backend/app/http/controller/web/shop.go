package web

import (
	"douyin-backend/app/global/consts"
	"douyin-backend/app/model/shop"
	"douyin-backend/app/utils/auth"
	"douyin-backend/app/utils/response"
	"github.com/gin-gonic/gin"
)

// ShopController 处理商品相关的 HTTP 接口。
type ShopController struct {
}

// GetShopRecommended 获取推荐商品列表。
func (u *ShopController) GetShopRecommended(ctx *gin.Context) {
	var uid = auth.GetUidFromToken(ctx)
	var pageNo = ctx.GetFloat64(consts.ValidatorPrefix + "pageNo")
	var pageSize = ctx.GetFloat64(consts.ValidatorPrefix + "pageSize")
	list, total, ok := shop.CreateShopFactory("").GetShopRecommended(uid, int64(pageNo), int64(pageSize))
	if !ok {
		response.Fail(ctx, consts.CurdSelectFailCode, consts.CurdSelectFailMsg, "获取推荐商品失败")
	} else {
		response.Success(ctx, consts.CurdStatusOkMsg, gin.H{
			"total": total,
			"list":  list,
		})
	}
}
