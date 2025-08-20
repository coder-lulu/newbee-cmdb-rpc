package citypeattributegroupitem

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeAttributeGroupItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeAttributeGroupItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeAttributeGroupItemLogic {
	return &CreateCiTypeAttributeGroupItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeAttributeGroupItemLogic) CreateCiTypeAttributeGroupItem(in *cmdb.CiTypeAttributeGroupItemInfo) (*cmdb.BaseIDResp, error) {
	result, err := l.svcCtx.DB.CiTypeAttributeGroupItem.Create().
		SetNotNilSort(in.Sort).
		SetNotNilGroupID(in.GroupId).
		SetNotNilAttrID(in.AttrId).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
