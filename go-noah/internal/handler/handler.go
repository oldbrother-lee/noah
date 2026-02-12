package handler

import (
	"go-noah/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// GetUserIdFromCtx 从 context 中获取用户 ID（工具函数）
func GetUserIdFromCtx(ctx *gin.Context) uint {
	v, exists := ctx.Get("claims")
	if !exists {
		return 0
	}
	return v.(*jwt.MyCustomClaims).UserId
}
