package middleware

import (
	"go-noah/api"
	"go-noah/pkg/jwt"
	"go-noah/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			logger.WithContext(ctx).Warn("No token", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}))
			api.HandleError(ctx, http.StatusUnauthorized, api.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			logger.WithContext(ctx).Error("token error", zap.Any("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}), zap.Error(err))
			api.HandleError(ctx, http.StatusUnauthorized, api.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

func NoStrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			tokenString, _ = ctx.Cookie("accessToken")
		}
		if tokenString == "" {
			tokenString = ctx.Query("accessToken")
		}
		if tokenString == "" {
			ctx.Next()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

func recoveryLoggerFunc(ctx *gin.Context, logger *log.Logger) {
	if userInfo, ok := ctx.MustGet("claims").(*jwt.MyCustomClaims); ok {
		logger.WithValue(ctx, zap.Any("UserId", userInfo.UserId))
	}
}
