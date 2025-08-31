package citypegroup

import (
	"context"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/utils/pointy"
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
