package auth

import (
	"context"
	"redis-demo/internal/dao"
	jwtLogic "redis-demo/internal/logic/jwt"
	"redis-demo/internal/model/do"
	"redis-demo/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"golang.org/x/crypto/bcrypt"
)

// Register 创建新用户，并返回用户ID。
// 注册时会检查用户名唯一性，并将密码转换为 bcrypt 哈希后保存。
func Register(ctx context.Context, username, password, nickname string) (userId uint64, err error) {
	// 检查用户名是否已存在
	count, err := dao.User.Ctx(ctx).Where("username", username).Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("用户名已存在")
	}

	//使用bcrypt加密密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// 插入新用户记录
	result, err := dao.User.Ctx(ctx).Data(do.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		Nickname:     nickname,
		Status:       1,
	}).Insert()
	if err != nil {
		return 0, err
	}

	// 返回插入的用户ID
	insertId, err := result.LastInsertId() // 获取最后插入的ID
	if err != nil {
		return 0, err
	}
	return uint64(insertId), nil
}

// Login 校验用户账号密码，成功后返回 JWT。
// 这里不直接暴露“用户不存在”或“密码错误”的具体原因，避免泄露账号是否存在。
func Login(ctx context.Context, username, password, captcha string) (string, error) {
	var user entity.User

	// 先校验验证码，验证码错误时不继续校验账号密码。
	ok, err := VerifyCaptcha(ctx, username, captcha)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", gerror.New("验证码错误")
	}

	//查询用户
	err = dao.User.Ctx(ctx).Where("username", username).Scan(&user)
	if err != nil {
		return "", gerror.New("用户名或密码错误")
	}
	//用户不存在或账号被禁用
	if user.Id == 0 {
		return "", gerror.New("用户名或密码错误")
	}
	if user.Status != 1 {
		return "", gerror.New("账号已禁用")
	}
	//验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", gerror.New("用户名或密码错误")
	}

	// 验证码已经通过，登录成功后删除验证码，保证验证码只能使用一次。
	if err := DeleteCaptcha(ctx, username); err != nil {
		return "", err
	}

	token, err := jwtLogic.GenerateToken(ctx, int64(user.Id), user.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}
