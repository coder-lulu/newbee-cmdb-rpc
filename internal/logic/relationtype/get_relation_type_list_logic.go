package relationtype

import (
	"context"
	"time"

	"gitee.com/link234/cmdb-rpc/ent/predicate"
	"gitee.com/link234/cmdb-rpc/ent/relationtype"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRelationTypeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRelationTypeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRelationTypeListLogic {
	return &GetRelationTypeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRelationTypeListLogic) GetRelationTypeList(in *cmdb.RelationTypeListReq) (*cmdb.RelationTypeListResp, error) {
	var predicates []predicate.RelationType
	if in.CreatedAt != nil {
		predicates = append(predicates, relationtype.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, relationtype.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, relationtype.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.Name != nil {
		predicates = append(predicates, relationtype.NameContains(*in.Name))
	}
	result, err := l.svcCtx.DB.RelationType.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.RelationTypeListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		resp.Data = append(resp.Data, &cmdb.RelationTypeInfo{
			Id:        &v.ID,
			CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Name:      &v.Name,
			Code:      &v.Code,
			Category:  pointy.GetPointer(string(v.Category)),
			Direction: pointy.GetPointer(string(v.Direction)),
		})
	}

	return resp, nil
}
