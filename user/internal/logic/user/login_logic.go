// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"
	"time"

	"github.com/mirai-zen/forge/user/internal/model"
	"github.com/mirai-zen/forge/user/internal/svc"
	"github.com/mirai-zen/forge/user/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	var user model.User
	if err := l.svcCtx.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		l.Errorf("user not found: %s", req.Username)
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		l.Errorf("password mismatch for user: %s", req.Username)
		return nil, errors.New("invalid credentials")
	}

	expiresAt := time.Now().Add(time.Duration(l.svcCtx.Config.JWT.Expire) * time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.Id,
		"username": user.Username,
		"role":     user.Role,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	})

	signedToken, err := token.SignedString([]byte(l.svcCtx.Config.JWT.Secret))
	if err != nil {
		l.Errorf("jwt sign failed: %v", err)
		return nil, errors.New("internal error")
	}

	l.Infof("user logged in: %s (id=%d)", user.Username, user.Id)
	return &types.LoginResp{
		Token:     signedToken,
		UserId:    user.Id,
		Username:  user.Username,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}
