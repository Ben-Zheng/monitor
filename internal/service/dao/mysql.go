package dao

import (
	"context"
	"fmt"
	"monitor/config"
	"time"

	"gorm.io/gorm/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	globalDB  *gorm.DB
	injectors = make([]func(*daoInit), 0)
)

// return &config.MySQLConfg

type daoInit struct {
	db          *gorm.DB
	autoMigrate bool
}

func Connect(cfg *config.DBConfig) (*gorm.DB, error) {
	if cfg == nil {
		panic(fmt.Errorf("mySQL config is nil"))
	}

	dsn := cfg.DSN
	dialector := mysql.Open(dsn)
	level := logger.Warn
	if cfg.Debug {
		level = logger.Info
	}
	newLogger := logger.Default.LogMode(level) // 设置日志级别为 Info，显示所有 SQL 语句

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	} else {
		sqlDB, err := db.DB()
		if err != nil {
			panic(err)
		}
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		callInjector(&daoInit{db: db, autoMigrate: cfg.AutoMigrate})
		registerTimeOutCallbacks(db)
		globalDB = db
	}

	hlog.Info("mysql db connected success")
	return db, nil
}

func callInjector(d *daoInit) {
	for _, v := range injectors {
		v(d)
	}
}

// 如果使用跨模型事务则传参.
func GetDB(tx ...*gorm.DB) *gorm.DB {
	var db *gorm.DB

	if len(tx) == 0 {
		db = globalDB
	} else {
		db = tx[0]
	}

	return db
}

func registerTimeOutCallbacks(db *gorm.DB) {
	query := db.Callback().Query()
	query.Before("gorm:query").Register("custom:before_gorm_query", GormBefore)
	query.After("gorm:after_query").Register("custom:after_gorm_query", GormAfter)
	del := db.Callback().Delete()
	del.Before("gorm:delete").Register("custom:before_gorm_delete", GormBefore)
	del.After("gorm:after_delete").Register("custom:after_gorm_delete", GormAfter)
	update := db.Callback().Update()
	update.Before("gorm:update").Register("custom:before_gorm_update", GormBefore)
	update.After("gorm:after_update").Register("custom:after_gorm_update", GormAfter)
	create := db.Callback().Create()
	create.Before("gorm:create").Register("custom:before_gorm_create", GormBefore)
	create.After("gorm:after_create").Register("custom:after_gorm_create", GormAfter)
}

func GormBefore(db *gorm.DB) {
	var ctxInfo *CtxInfo
	ctxInfo = getCtxInfo(db.Statement.Context)
	if ctxInfo == nil {
		ctxInfo = &CtxInfo{OriginContext: db.Statement.Context}
	}
	var cancelFunc context.CancelFunc
	db.Statement.Context, cancelFunc = context.WithTimeout(ctxInfo.OriginContext, 180*time.Second)
	ctxInfo.cancelFunc = &cancelFunc
	const ctxInfoKey CtxInfoKeyType = "ctxInfo"
	db.Statement.Context = context.WithValue(
		db.Statement.Context,
		ctxInfoKey,
		ctxInfo,
	)
}

func GormAfter(db *gorm.DB) {
	ctxInfo := getCtxInfo(db.Statement.Context)
	if ctxInfo != nil {
		f := *ctxInfo.cancelFunc
		f()
	}
}

// CtxInfo
type CtxInfo struct {
	cancelFunc    *context.CancelFunc
	OriginContext context.Context
}

type CtxInfoKeyType string

// getCtxInfo get ctx info from context
func getCtxInfo(ctx context.Context) *CtxInfo {
	if ctxInfo, ok := ctx.Value("ctxInfo").(*CtxInfo); ok {
		return ctxInfo
	}
	return nil
}

type TableModel interface {
	TableName() string
}

func registerInjector(f func(*daoInit)) {
	injectors = append(injectors, f)
}

// 自动初始化表结构.
func setupTableModel(d *daoInit, model TableModel) {
	if d.autoMigrate {
		err := d.db.AutoMigrate(model)
		if err != nil {
			panic(err)
		}
	}
}
