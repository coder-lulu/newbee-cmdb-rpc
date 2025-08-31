package citype

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citype"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypegroupitem"
	"github.com/coder-lulu/newbee-cmdb-rpc/ent/predicate"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	uuid_helper "github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/uuid"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-common/utils/uuidx"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeListLogic {
	return &GetCiTypeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeListLogic) GetCiTypeList(in *cmdb.CiTypeListReq) (*cmdb.CiTypeListResp, error) {
	// 如果指定了group_id，则通过CiTypeGroupItem查询
	if in.GroupId != nil {
		return l.getCiTypeListByGroupId(*in.GroupId, in)
	}

	// 否则直接查询CiType表
	var predicates []predicate.CiType

	if in.Status != nil {
		predicates = append(predicates, citype.StatusEQ(uint8(*in.Status)))
	}
	if in.Sort != nil {
		predicates = append(predicates, citype.SortEQ(*in.Sort))
	}
	if in.Name != nil {
		predicates = append(predicates, citype.NameContains(*in.Name))
	}
	if in.Alias != nil {
		predicates = append(predicates, citype.AliasContains(*in.Alias))
	}
	if in.UniqueId != nil {
		predicates = append(predicates, citype.UniqueIDEQ(*in.UniqueId))
	}
	if in.CreatedBy != nil {
		predicates = append(predicates, citype.CreatedByEQ(uuidx.ParseUUIDString(*in.CreatedBy)))
	}
	if in.Icon != nil {
		predicates = append(predicates, citype.IconContains(*in.Icon))
	}
	if in.DefaultOrderAttr != nil {
		predicates = append(predicates, citype.DefaultOrderAttrIDEQ(*in.DefaultOrderAttr))
	}
	if in.ShowId != nil {
		predicates = append(predicates, citype.ShowIDEQ(*in.ShowId))
	}

	result, err := l.svcCtx.DB.CiType.Query().Where(predicates...).All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &cmdb.CiTypeListResp{}
	resp.Total = uint64(len(result))

	for _, v := range result {
		var uniqueConst []*cmdb.CiTypeUniqueConst
		if len(v.UniqueConst) > 0 {
			for _, v := range v.UniqueConst {
				uniqueConst = append(uniqueConst, &cmdb.CiTypeUniqueConst{
					AttrIds: v.AttrIds,
				})
			}
		}

		resp.Data = append(resp.Data, &cmdb.CiTypeInfo{
			Id:               &v.ID,
			CreatedAt:        pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:        pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:           pointy.GetPointer(uint32(v.Status)),
			Sort:             &v.Sort,
			Name:             &v.Name,
			Alias:            &v.Alias,
			UniqueId:         &v.UniqueID,
			IsInherited:      v.IsInherited,
			CreatedBy:        uuid_helper.SafeUUIDToStringPtr(v.CreatedBy),
			Icon:             v.Icon,
			DefaultOrderAttr: v.DefaultOrderAttrID,
			ShowId:           v.ShowID,
			UniqueConst:      uniqueConst,
		})
	}

	return resp, nil
}

// getCiTypeListByGroupId 根据group_id查询CiType列表
func (l *GetCiTypeListLogic) getCiTypeListByGroupId(groupId uint64, in *cmdb.CiTypeListReq) (*cmdb.CiTypeListResp, error) {
	var predicates []predicate.CiTypeGroupItem
	predicates = append(predicates, citypegroupitem.GroupIDEQ(groupId))

	result, err := l.svcCtx.DB.CiTypeGroupItem.Query().WithCiType().Where(predicates...).All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	ciTypes := make([]*cmdb.CiTypeInfo, 0)
	if len(result) == 0 {
		return &cmdb.CiTypeListResp{
			Total: 0,
			Data:  ciTypes,
		}, nil
	}

	for _, v := range result {
		ciType, err := l.svcCtx.DB.CiType.Get(l.ctx, v.TypeID)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		// 处理其他查询条件
		if in.Status != nil && ciType.Status != uint8(*in.Status) {
			continue
		}
		if in.Name != nil && !contains(ciType.Name, *in.Name) {
			continue
		}
		if in.Alias != nil && !contains(ciType.Alias, *in.Alias) {
			continue
		}
		if in.UniqueId != nil && ciType.UniqueID != *in.UniqueId {
			continue
		}
		if in.CreatedBy != nil && ciType.CreatedBy != nil && ciType.CreatedBy.String() != *in.CreatedBy {
			continue
		}
		if in.Icon != nil && ciType.Icon != nil && !contains(*ciType.Icon, *in.Icon) {
			continue
		}
		if in.DefaultOrderAttr != nil && ciType.DefaultOrderAttrID != nil && *ciType.DefaultOrderAttrID != *in.DefaultOrderAttr {
			continue
		}
		if in.ShowId != nil && ciType.ShowID != nil && *ciType.ShowID != *in.ShowId {
			continue
		}

		// 处理唯一性约束
		var uniqueConst []*cmdb.CiTypeUniqueConst
		if len(ciType.UniqueConst) > 0 {
			for _, v := range ciType.UniqueConst {
				uniqueConst = append(uniqueConst, &cmdb.CiTypeUniqueConst{
					AttrIds: v.AttrIds,
				})
			}
		}

		ciTypes = append(ciTypes, &cmdb.CiTypeInfo{
			Id:               &ciType.ID,
			CreatedAt:        pointy.GetPointer(ciType.CreatedAt.UnixMilli()),
			UpdatedAt:        pointy.GetPointer(ciType.UpdatedAt.UnixMilli()),
			Status:           pointy.GetPointer(uint32(ciType.Status)),
			Sort:             &v.Sort,
			Name:             &ciType.Name,
			Alias:            &ciType.Alias,
			UniqueId:         &ciType.UniqueID,
			IsInherited:      ciType.IsInherited,
			CreatedBy:        uuid_helper.SafeUUIDToStringPtr(ciType.CreatedBy),
			Icon:             ciType.Icon,
			DefaultOrderAttr: ciType.DefaultOrderAttrID,
			ShowId:           ciType.ShowID,
			UniqueConst:      uniqueConst,
		})
	}

	return &cmdb.CiTypeListResp{
		Total: uint64(len(ciTypes)),
		Data:  ciTypes,
	}, nil
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
