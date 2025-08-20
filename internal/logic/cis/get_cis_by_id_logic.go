package cis

import (
	"context"
	"fmt"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCisByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCisByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCisByIdLogic {
	return &GetCisByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCisByIdLogic) GetCisById(in *cmdb.IDReq) (*cmdb.CisInfo, error) {
	// 查询CI实例基础信息
	result, err := l.svcCtx.DB.Cis.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 加载动态属性值
	attributes, err := LoadCiAttributes(l.ctx, l.svcCtx.DB, in.Id, nil, nil)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 转换为响应格式
	cisInfo := CisEntToProto(result, attributes)
	if cisInfo == nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger,
			fmt.Errorf("failed to convert CI entity to proto"), in)
	}

	return cisInfo, nil
}
