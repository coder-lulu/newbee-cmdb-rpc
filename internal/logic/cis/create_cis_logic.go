package cis

import (
	"context"

	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/coder-lulu/newbee-common/v2/msg/errormsg"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCisLogic {
	return &CreateCisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCisLogic) CreateCis(in *cmdb.CisInfo) (*cmdb.BaseIDResp, error) {
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

	// 校验属性唯一性（创建时不需要排除任何CI）
	if len(in.Attributes) > 0 && in.TypeId != nil {
		err = ValidateUniqueAttributes(l.ctx, l.svcCtx.DB, *in.TypeId, in.Attributes, nil)
		if err != nil {
			tx.Rollback()
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 创建CI实例基础信息
	cisBuilder := tx.Cis.Create()
	cisBuilder = CisCreateBuilderSetter(cisBuilder, in)

	result, err := cisBuilder.Save(l.ctx)
	if err != nil {
		tx.Rollback()
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 保存动态属性值
	if len(in.Attributes) > 0 {
		err = SaveCiAttributes(l.ctx, tx, result.ID, in.Attributes)
		if err != nil {
			tx.Rollback()
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 处理关系数据
	if in.Relations != nil {
		err = ProcessCiRelations(l.ctx, tx, result.ID, in.Relations, l.Logger)
		if err != nil {
			tx.Rollback()
			if ce, ok := err.(*errorx.CodeError); ok {
				return nil, ce
			}
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &cmdb.BaseIDResp{Id: result.ID, Msg: errormsg.CreateSuccess}, nil
}
