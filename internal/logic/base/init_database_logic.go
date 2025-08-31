package base

import (
	"context"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/coder-lulu/newbee-common/msg/errormsg"
	"github.com/coder-lulu/newbee-common/msg/logmsg"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-cmdb-rpc/ent"
	"github.com/coder-lulu/newbee-cmdb-rpc/internal/svc"
	"github.com/coder-lulu/newbee-cmdb-rpc/types/cmdb"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitDatabaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitDatabaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitDatabaseLogic {
	return &InitDatabaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InitDatabaseLogic) InitDatabase(in *cmdb.Empty) (*cmdb.BaseResp, error) {
	if err := l.svcCtx.DB.Schema.Create(l.ctx, schema.WithForeignKeys(false)); err != nil {
		logx.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
		return nil, errorx.NewInternalError(err.Error())
	}

	// errHandler := func(err error) (*cmdb.BaseResp, error) {
	// 	logx.Errorw(logmsg.DatabaseError, logx.Field("detail", err.Error()))
	// 	return nil, errorx.NewInternalError(err.Error())
	// }

	// // 插入初始数据
	// err := l.InsertInitData()
	// if err != nil {
	// 	return errHandler(err)
	// }

	return &cmdb.BaseResp{Msg: errormsg.Success}, nil
}

func (l *InitDatabaseLogic) InsertInitData() error {
	//插入默认分组
	var groups []*ent.CiTypeGroupCreate
	groups = append(groups, l.svcCtx.DB.CiTypeGroup.Create().SetName("其它").SetDescription("其它").SetIcon("icon-other").SetSort(999))
	groups = append(groups, l.svcCtx.DB.CiTypeGroup.Create().SetName("基类").SetDescription("基类").SetIcon("icon-base").SetSort(998))

	err := l.svcCtx.DB.CiTypeGroup.CreateBulk(groups...).Exec(l.ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	// 关系类型
	relationTypes := []*ent.RelationTypeCreate{
		l.svcCtx.DB.RelationType.Create().SetName("包含").SetCode("contain").SetCategory("logic").SetDirection("bidirectional"),
		l.svcCtx.DB.RelationType.Create().SetName("部署").SetCode("deploy").SetCategory("logic").SetDirection("bidirectional"),
		l.svcCtx.DB.RelationType.Create().SetName("安装").SetCode("install").SetCategory("logic").SetDirection("bidirectional"),
		l.svcCtx.DB.RelationType.Create().SetName("关联").SetCode("associate").SetCategory("logic").SetDirection("bidirectional"),
	}
	err = l.svcCtx.DB.RelationType.CreateBulk(relationTypes...).Exec(l.ctx)
	if err != nil {
		return errorx.NewInternalError(err.Error())
	}

	return nil
}
