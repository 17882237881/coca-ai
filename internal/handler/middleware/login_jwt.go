package middleware

import (
	"coca-ai/pkg/jwtx"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type LoginJWTMiddleware struct {
	jwtHandler *jwtx.JWTHandler
	redisCmd   redis.Cmdable
}

func NewLoginJWTMiddleware(jwtHandler *jwtx.JWTHandler, redisCmd redis.Cmdable) *LoginJWTMiddleware {
	return &LoginJWTMiddleware{
		jwtHandler: jwtHandler,
		redisCmd:   redisCmd,
	}
}

// Check 检查登录状态
func (m *LoginJWTMiddleware) Check() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := m.extractToken(ctx)
		if tokenStr == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtHandler.ParseToken(tokenStr)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 2. 检查 Redis 黑名单 (SSID)
		if claims.Ssid != "" {
			err := m.redisCmd.Get(ctx, "users:ssid:"+claims.Ssid).Err()
			if err == nil {
				// Key 存在，说明在黑名单里 -> 拒绝
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			// 如果是 redis.Nil 说明不在黑名单，通过
			// 如果是其他错误，也拒绝? 为了安全，Redis挂了最好拒绝访问，或者降级通过。这里选择 Fail-Secure (拒绝)
			if err != redis.Nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		// 将 User Claims 存入 Context，后续 Handle 可以直接取用
		ctx.Set("claims", claims)
		ctx.Set("uid", claims.Uid)

		ctx.Next()
	}
}

func (m *LoginJWTMiddleware) extractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return ""
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}
