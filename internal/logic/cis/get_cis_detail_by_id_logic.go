package cis

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/cis"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCisDetailByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCisDetailByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCisDetailByIdLogic {
	return &GetCisDetailByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCisDetailByIdLogic) GetCisDetailById(in *cmdb.IDReq) (*cmdb.CisDetailInfo, error) {
	// 查询CI实例基础信息，包含CI类型信息
	result, err := l.svcCtx.DB.Cis.Query().
		Where(cis.IDEQ(in.Id)).
		WithCiType().
		Only(l.ctx)
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

	// 构建详情响应
	detailInfo := &cmdb.CisDetailInfo{
		CiInfo: cisInfo,
	}

	// 设置CI类型信息
	if result.Edges.CiType != nil {
		detailInfo.CiTypeName = result.Edges.CiType.Name
		detailInfo.CiTypeAlias = result.Edges.CiType.Alias
	}

	return detailInfo, nil
}
