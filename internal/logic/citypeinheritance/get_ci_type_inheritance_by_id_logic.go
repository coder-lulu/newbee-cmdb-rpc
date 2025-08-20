package citypeinheritance

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeInheritanceByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeInheritanceByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeInheritanceByIdLogic {
	return &GetCiTypeInheritanceByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeInheritanceByIdLogic) GetCiTypeInheritanceById(in *cmdb.IDReq) (*cmdb.CiTypeInheritanceInfo, error) {
	result, err := l.svcCtx.DB.CiTypeInheritance.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.CiTypeInheritanceInfo{
		Id:        &result.ID,
		CreatedAt: pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt: pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		ParentId:  &result.ParentID,
		ChildId:   &result.ChildID,
	}, nil
}
