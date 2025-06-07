package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/haomiao000/raftchain/library/model"
	"github.com/haomiao000/raftchain/library/query"
	"github.com/haomiao000/raftchain/library/utils"

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

	// 登录成功，生成一个有效期为24小时的token
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

// ChangePassword 验证旧信息并更新为新密码
func ChangePassword(ctx context.Context, username, email, oldPassword, newPassword string) error {
	// 1. 获取用户信息
	user, err := GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户名、邮箱或原密码错误")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 2. 验证邮箱是否匹配
	if user.Email != email {
		return errors.New("用户名、邮箱或原密码错误")
	}

	// 3. 验证原密码是否正确
	expectedOldPasswordHash := utils.GenSha256(oldPassword)
	if user.Password != expectedOldPasswordHash {
		return errors.New("用户名、邮箱或原密码错误")
	}

	// 4. 更新为新密码
	newPasswordHash := utils.GenSha256(newPassword)
	u := query.User
	result, err := u.WithContext(ctx).Where(u.ID.Eq(user.ID)).Update(u.Password, newPasswordHash)
	if err != nil {
		return fmt.Errorf("数据库更新密码失败: %w", err)
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到用户或无需更新")
	}

	return nil
}