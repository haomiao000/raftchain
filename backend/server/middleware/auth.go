package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/haomiao000/raftchain/library/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// 解析token
func ParseToken(tokenString string) (*utils.MyClaims, error) {
	var claims utils.MyClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, utils.GetKey)

	if err != nil {
		// 更细致的错误判断
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("token 格式错误")
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, errors.New("token 已过期或尚未生效")
			} else {
				return nil, fmt.Errorf("无法处理 token: %v", err)
			}
		}
		return nil, fmt.Errorf("无法解析 token: %v", err)
	}
	if !token.Valid {
		return nil, errors.New("token 无效")
	}
	return &claims, nil
}

// Token 验证中间件
func VerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请求未包含授权 Token (Authorization header is missing)"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "授权 Token 格式错误 (must be Bearer token)"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许的源，根据你的前端端口设置
		// 为了更安全，生产环境中应指定确切的前端域名
		// 如果你的前端可能通过 localhost:3000 或 127.0.0.1:3000 访问，可以都加上，或用更灵活的配置
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:3000")
		// 如果你的前端也可能用 localhost:3000 访问，并且你想同时允许：
		// origin := c.Request.Header.Get("Origin")
		// if origin == "http://127.0.0.1:3000" || origin == "http://localhost:3000" {
		//  c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		// }

		// 是否允许发送 Cookie。如果你的认证依赖于 Cookie，则设为 true。
		// 对于 JWT Token 通过 Authorization Header 传递的情况，这个通常不是必须的，除非你有其他用途。
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// 允许的请求头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")

		// 允许的 HTTP 方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// 处理预检请求 (OPTIONS)
		// 当浏览器发送带有自定义头（如 Authorization）的非简单请求（如 POST json）时，会先发一个 OPTIONS 请求。
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent) // 204 No Content
			return
		}

		c.Next() // 继续处理请求链
	}
}