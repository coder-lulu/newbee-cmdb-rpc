package citypeattribute

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/logic/attribute"
	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"
	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCiTypeAttributeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCiTypeAttributeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCiTypeAttributeLogic {
	return &CreateCiTypeAttributeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCiTypeAttributeLogic) CreateCiTypeAttribute(in *cmdb.CiTypeAttributeItem) (*cmdb.BaseResp, error) {
	// 启用事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	attr, err := attribute.NewCreateAttributeLogic(l.ctx, l.svcCtx).CreateAttribute(in.Attribute)
	if err != nil {
		return nil, err
	}
	// 如果类型ID和分组ID都为0，则是创建属性
	if in.TypeId == 0 && *in.GroupId == 0 {
		return &cmdb.BaseResp{Msg: errormsg.CreateSuccess}, nil
	}

	_, err = tx.CiTypeAttribute.Create().
		SetNotNilTypeID(&in.TypeId).
		SetNotNilAttrID(&attr.Id).
		SetNotNilIsRequired(in.IsRequired).
		SetNotNilListShow(in.ListShow).
		SetNotNilIsEdit(in.IsEdit).
		SetNotNilDetailShow(in.DetailShow).
		SetNotNilIsUnique(in.IsUnique).
		// 其他字段按需补充
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	_, err = tx.CiTypeAttributeGroupItem.Create().
		SetNotNilGroupID(in.GroupId).
		SetNotNilAttrID(&attr.Id).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 8. 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &cmdb.BaseResp{Msg: errormsg.CreateSuccess}, nil
}
