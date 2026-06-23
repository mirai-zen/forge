// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"
	"errors"

	"github.com/mirai-zen/forge/user/internal/model"
	"github.com/mirai-zen/forge/user/internal/svc"
	"github.com/mirai-zen/forge/user/internal/types"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyTokenLogic {
	return &VerifyTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyTokenLogic) VerifyToken(req *types.VerifyTokenReq) (resp *types.VerifyTokenResp, err error) {
	token, err := jwt.Parse(req.Token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(l.svcCtx.Config.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return &types.VerifyTokenResp{Valid: false}, nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &types.VerifyTokenResp{Valid: false}, nil
	}

	userIdFloat, _ := claims["user_id"].(float64)
	userId := int64(userIdFloat)

	// 确认用户仍然存在且未被禁用
	var user model.User
	if err := l.svcCtx.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		return &types.VerifyTokenResp{Valid: false}, nil
	}

	username, _ := claims["username"].(string)
	return &types.VerifyTokenResp{
		Valid:    true,
		UserId:   userId,
		Username: username,
	}, nil
}
