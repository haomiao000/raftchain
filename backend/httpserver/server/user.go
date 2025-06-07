package server

import (
	"net/http"

	"github.com/haomiao000/raftchain/backend/server/controller/user"

	"github.com/gin-gonic/gin"
)

// Register .
// @router /api/v2/user/register [POST]
func Register(c *gin.Context) {
	type param struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Password string `json:"password" binding:"required,min=6,max=100"`
		Email    string `json:"email" binding:"omitempty,email"`
	}
	var p param
	err := c.ShouldBindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}
	id, err := user.Register(c.Request.Context(), p.Username, p.Password, p.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "用户注册成功",
		"user_id":  id,
		"username": p.Username,
		"email":    p.Email,
	})
}

// Login .
// @router /api/v2/user/login [POST]
func Login(c *gin.Context) {
	type param struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Password string `json:"password" binding:"required,min=6,max=100"`
	}
	var p param
	err := c.ShouldBindJSON(&p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}
	resp, err := user.Login(c.Request.Context(), p.Username, p.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "登陆成功",
		"data": resp,
	})
}

// ChangePassword .
// @router /api/v2/user/change-password [POST]
func ChangePassword(c *gin.Context) {
	type param struct {
		Username    string `json:"username" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required"`       // 原密码
		NewPassword string `json:"new_password" binding:"required,min=6"` // 新密码
	}
	var p param
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	err := user.ChangePassword(c.Request.Context(), p.Username, p.Email, p.Password, p.NewPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码重置成功",
	})
}