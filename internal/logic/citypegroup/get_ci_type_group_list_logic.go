package citypegroup

import (
	"context"
	"gitee.com/link234/cmdb-rpc/ent/citypegroup"
	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeGroupListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeGroupListLogic {
	return &GetCiTypeGroupListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeGroupListLogic) GetCiTypeGroupList(in *cmdb.CiTypeGroupListReq) (*cmdb.CiTypeGroupListResp, error) {
	var predicates []predicate.CiTypeGroup
	if in.Name != nil {
		predicates = append(predicates, citypegroup.NameContains(*in.Name))
	}
	result, err := l.svcCtx.DB.CiTypeGroup.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeGroupListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.CiTypeGroupInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Sort:      &v.Sort,
			Name:      &v.Name,
		})
	}

	return resp, nil
}
