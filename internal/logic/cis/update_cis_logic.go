package cis

import (
	"context"

	"gitee.com/link234/cmdb-rpc/internal/svc"
	"gitee.com/link234/cmdb-rpc/internal/utils/dberrorhandler"
	"gitee.com/link234/cmdb-rpc/types/cmdb"

	"gitee.com/link234/newbee-backend-common/msg/errormsg"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCisLogic {
	return &UpdateCisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCisLogic) UpdateCis(in *cmdb.CisInfo) (*cmdb.BaseResp, error) {
	// 开启事务
	tx, err := l.svcCtx.DB.Tx(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// 获取当前CI的类型ID用于唯一性校验
	var typeID uint64
	if in.TypeId != nil {
		typeID = *in.TypeId
	} else {
		// 如果请求中没有TypeId，需要从数据库中获取
		currentCi, err := l.svcCtx.DB.Cis.Get(l.ctx, *in.Id)
		if err != nil {
			tx.Rollback()
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		typeID = currentCi.TypeID
	}

	// 校验属性唯一性（更新时需要排除当前CI）
	if len(in.Attributes) > 0 {
		err = ValidateUniqueAttributes(l.ctx, l.svcCtx.DB, typeID, in.Attributes, in.Id)
		if err != nil {
			tx.Rollback()
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 更新CI实例基础信息
	cisBuilder := tx.Cis.UpdateOneID(*in.Id)
	cisBuilder = CisUpdateBuilderSetter(cisBuilder, in)

	_, err = cisBuilder.Save(l.ctx)
	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 更新动态属性值
	if len(in.Attributes) > 0 {
		err = SaveCiAttributes(l.ctx, tx, *in.Id, in.Attributes)
		if err != nil {
			tx.Rollback()
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseResp{Msg: errormsg.UpdateSuccess}, nil
}
