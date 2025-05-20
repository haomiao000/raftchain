package user

import (
	"context"
	"errors"
	"fmt"

	"main/library/model"
	"main/library/query"
	"main/library/utils"

	"gorm.io/gorm"
)

// 通过username获取user
func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	u := query.User
	user, err := u.WithContext(ctx).Where(u.Username.Eq(username)).First()
	if err != nil {
		return nil, err
	}
	return user, err
}

// 创建user
func CreateUserByInst(ctx context.Context, user *model.User) (int64, error) {
	u := query.User
	err := u.WithContext(ctx).Create(user)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// 注册
func Register(ctx context.Context, username string, password string, email string) (int64, error) {
	_, err := GetUserByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.New("注册用户已存在")
	}
	newUser := &model.User{
		ID:       0,
		Username: username,
		Password: utils.GenSha256(password),
		Email:    email,
	}
	id, err := CreateUserByInst(ctx, newUser)
	if err != nil {
		return 0, err
	}
	return id, err
}

// 登陆
func Login(ctx context.Context, username string, password string) (any, error) {
	user, err := GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("登录时获取用户信息失败: %w", err) // Other DB error
	}

	expectedPasswordHash := utils.GenSha256(password)
	if user.Password != expectedPasswordHash {
		return nil, errors.New("用户名或密码错误")
	}

	tokenString, err := utils.GenToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("生成认证令牌失败: %w", err)
	}

	return map[string]interface{}{
		"token":    tokenString,
		"user_id":  user.ID,
		"username": user.Username,
	}, nil
}
