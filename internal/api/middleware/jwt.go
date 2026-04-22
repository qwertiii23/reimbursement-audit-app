package middleware

import (
	"strings"

	"reimbursement-audit/internal/pkg/crypto"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey   = "user_id"
	UserNameKey = "user_name"
	UserRoleKey = "user_role"
)

var noAuthPaths = map[string]bool{
	"/api/v1/auth/login":       true,
	"/health":                  true,
	"/ready":                   true,
	"/version":                 true,
	"/api/v1/knowledge/upload": true,
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if noAuthPaths[path] {
			c.Next()
			return
		}

		if strings.HasPrefix(path, "/uploads/") || strings.HasPrefix(path, "/api/v1/files/") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"code": 1002, "message": "未提供认证令牌"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(401, gin.H{"code": 1002, "message": "认证令牌格式错误"})
			c.Abort()
			return
		}

		claims, err := crypto.ParseToken(parts[1])
		if err != nil {
			c.JSON(401, gin.H{"code": 1002, "message": "认证令牌无效"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(401, gin.H{"code": 1002, "message": "认证令牌缺少用户信息"})
			c.Abort()
			return
		}

		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		c.Set(UserIDKey, userID)
		c.Set(UserNameKey, username)
		c.Set(UserRoleKey, role)

		c.Next()
	}
}
