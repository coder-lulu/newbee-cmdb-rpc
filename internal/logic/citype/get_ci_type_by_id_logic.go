package citype

import (
	"context"

	"errors"

	"gitee.com/link234/cmdb-rpc/ent/citype"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCiTypeByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCiTypeByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCiTypeByIdLogic {
	return &GetCiTypeByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCiTypeByIdLogic) GetCiTypeById(in *cmdb.IDReq) (*cmdb.CiTypeInfo, error) {
	result, err := l.svcCtx.DB.CiType.Query().Where(citype.IDEQ(in.Id)).WithParents().First(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if result == nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, errors.New("ci_type not found"), in)
	}

	var createdByStr *string
	if result.CreatedBy != nil {
		str := result.CreatedBy.String()
		createdByStr = &str
	}

	// 处理唯一性约束
	var uniqueConst []*cmdb.CiTypeUniqueConst
	if len(result.UniqueConst) > 0 {
		for _, v := range result.UniqueConst {
			uniqueConst = append(uniqueConst, &cmdb.CiTypeUniqueConst{
				AttrIds: v.AttrIds,
			})
		}
	}

	var inheritedModels []uint64
	if result.IsInherited != nil && *result.IsInherited && len(result.Edges.Parents) > 0 {
		for _, parent := range result.Edges.Parents {
			inheritedModels = append(inheritedModels, parent.ParentID)
		}
	}

	return &cmdb.CiTypeInfo{
		Id:               &result.ID,
		CreatedAt:        pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:        pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Status:           pointy.GetPointer(uint32(result.Status)),
		Sort:             &result.Sort,
		Name:             &result.Name,
		Alias:            &result.Alias,
		UniqueId:         &result.UniqueID,
		IsInherited:      result.IsInherited,
		CreatedBy:        createdByStr,
		Icon:             result.Icon,
		DefaultOrderAttr: result.DefaultOrderAttrID,
		ShowId:           result.ShowID,
		UniqueConst:      uniqueConst,
		InheritedModels:  inheritedModels,
	}, nil
}
