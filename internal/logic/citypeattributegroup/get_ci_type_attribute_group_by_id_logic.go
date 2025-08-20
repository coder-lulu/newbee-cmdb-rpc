package citypeattributegroup

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeAttributeGroupByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeAttributeGroupByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeAttributeGroupByIdLogic {
	return &GetCiTypeAttributeGroupByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeAttributeGroupByIdLogic) GetCiTypeAttributeGroupById(in *cmdb.IDReq) (*cmdb.CiTypeAttributeGroupInfo, error) {
	result, err := l.svcCtx.DB.CiTypeAttributeGroup.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeAttributeGroupInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Sort:      &result.Sort,
		Name:      &result.Name,
		TypeId:    &result.TypeID,
	}, nil
}
