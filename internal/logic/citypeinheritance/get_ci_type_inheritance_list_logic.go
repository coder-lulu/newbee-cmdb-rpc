package citypeinheritance

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeinheritance"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeInheritanceListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeInheritanceListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeInheritanceListLogic {
	return &GetCiTypeInheritanceListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeInheritanceListLogic) GetCiTypeInheritanceList(in *cmdb.CiTypeInheritanceListReq) (*cmdb.CiTypeInheritanceListResp, error) {
	var predicates []predicate.CiTypeInheritance
	if in.CreatedAt != nil {
		predicates = append(predicates, citypeinheritance.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, citypeinheritance.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, citypeinheritance.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.ParentId != nil {
		predicates = append(predicates, citypeinheritance.ParentIDEQ(*in.ParentId))
	}
	if in.ChildId != nil {
		predicates = append(predicates, citypeinheritance.ChildIDEQ(*in.ChildId))
	}
	result, err := l.svcCtx.DB.CiTypeInheritance.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeInheritanceListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.CiTypeInheritanceInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			ParentId:  &v.ParentID,
			ChildId:   &v.ChildID,
		})
	}

	return resp, nil
}
