// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"

	"github.com/mirai-zen/forge/user/internal/model"
	"github.com/mirai-zen/forge/user/internal/svc"
	"github.com/mirai-zen/forge/user/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	if len(req.Username) < 3 || len(req.Username) > 32 {
		return nil, errors.New("username must be 3-32 characters")
	}
	if len(req.Password) < 6 || len(req.Password) > 64 {
		return nil, errors.New("password must be 6-64 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("bcrypt hash failed: %v", err)
		return nil, errors.New("internal error")
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		Role:         "user",
	}

	if err := l.svcCtx.DB.Create(user).Error; err != nil {
		l.Errorf("create user failed: %v", err)
		return nil, errors.New("username already exists")
	}

	l.Infof("user registered: %s (id=%d)", user.Username, user.Id)
	return &types.RegisterResp{
		Id:      user.Id,
		Message: "ok",
	}, nil
}
