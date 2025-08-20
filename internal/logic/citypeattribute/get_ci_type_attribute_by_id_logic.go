package citypeattribute

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/logic/attribute"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeAttributeByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeAttributeByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeAttributeByIdLogic {
	return &GetCiTypeAttributeByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeAttributeByIdLogic) GetCiTypeAttributeById(in *cmdb.IDReq) (*cmdb.CiTypeAttributeItem, error) {
	result, err := l.svcCtx.DB.CiTypeAttribute.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 查询 Attribute 详情
	attr, err := attribute.NewGetAttributeByIdLogic(l.ctx, l.svcCtx).GetAttributeById(&cmdb.IDReq{Id: result.AttrID})
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	item, err := l.svcCtx.DB.CiTypeAttribute.Get(l.ctx, result.ID)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeAttributeItem{
		Id:                &item.ID,
		Attribute:         attr,
		TypeId:            result.TypeID,
		IsRequired:        &result.IsRequired,
		ListShow:          &result.ListShow,
		IsEdit:            &result.IsEdit,
		DetailShow:        &result.DetailShow,
		IsUnique:          &result.IsUnique,
		Sort:              &result.Sort,
		CiTypeAttributeId: &result.ID,
	}, nil
}
