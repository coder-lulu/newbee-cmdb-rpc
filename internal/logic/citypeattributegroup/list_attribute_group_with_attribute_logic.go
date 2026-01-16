package citypeattributegroup

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAttributeGroupWithAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAttributeGroupWithAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAttributeGroupWithAttributeLogic {
	return &ListAttributeGroupWithAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAttributeGroupWithAttributeLogic) ListAttributeGroupWithAttribute(in *cmdb.IDReq) (*cmdb.CiTypeAttributeGroupWithAttrListResp, error) {
	groups, err := l.svcCtx.DB.CiTypeAttributeGroup.Query().
		Where(citypeattributegroup.TypeIDEQ(in.Id)).
		Order(ent.Asc(citypeattributegroup.FieldSort)).
		WithGroupItems(func(q *ent.CiTypeAttributeGroupItemQuery) {
			q.Order(ent.Asc(citypeattributegroupitem.FieldSort))
			q.WithAttribute()
		}).
		All(l.ctx)

	if err != nil {
		return nil, err
	}

	resp := &cmdb.CiTypeAttributeGroupWithAttrListResp{}
	resp.Total = uint64(len(groups))

	for _, g := range groups {
		groupInfo := &cmdb.CiTypeAttributeGroupWithAttrInfo{
			Id:   &g.ID,
			Name: &g.Name,
			Sort: &g.Sort,
		}

		for _, item := range g.Edges.GroupItems {
			attr := item.Edges.Attribute
			if attr == nil {
				continue
			}

			// Map attribute to AttributeInfo
			attrInfo := &cmdb.AttributeInfo{
				Id:        &attr.ID,
				Name:      &attr.Name,
				Alias:     &attr.Alias,
				ValueType: pointy.GetPointer(string(attr.ValueType)),
				IsChoice:  &attr.IsChoice,
				IsList:    &attr.IsList,
				// Add other fields as needed
			}

			groupItemInfo := &cmdb.CiTypeAttributeGroupItemWithAttrInfo{
				Id:        &item.ID,
				GroupId:   &item.GroupID,
				AttrId:    &item.AttrID,
				Sort:      &item.Sort,
				Attribute: attrInfo,
			}
			groupInfo.Items = append(groupInfo.Items, groupItemInfo)
		}
		resp.Data = append(resp.Data, groupInfo)
	}

	return resp, nil
}
