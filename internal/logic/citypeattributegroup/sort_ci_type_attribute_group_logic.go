package citypeattributegroup

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent/citypeattributegroup"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SortCiTypeAttributeGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSortCiTypeAttributeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SortCiTypeAttributeGroupLogic {
	return &SortCiTypeAttributeGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SortCiTypeAttributeGroupLogic) SortCiTypeAttributeGroup(in *cmdb.CiTypeAttributeGroupSortReq) (*cmdb.BaseResp, error) {
	// 增加事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, v := range in.Data {
		if v.Name == "其它" {
			continue
		}
		err := tx.CiTypeAttributeGroup.Update().SetSort(uint32(v.Sort)).Where(citypeattributegroup.ID(v.GroupId)).Exec(l.ctx)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseResp{}, nil
}
