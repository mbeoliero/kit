package connector

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mbeoliero/kit/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"gorm.io/plugin/opentelemetry/tracing"
)

func MustInitGorm(cfg MysqlConfig) *gorm.DB {
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 3
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 30
	}

	db, err := InitGorm(cfg)
	if err != nil {
		log.Error("MustInitGorm init db err %+v", err)
		panic(err)
	}
	if db != nil {
		db = AddTraceLogger(db, cfg.DisableLog)
	}
	return db
}

func InitGorm(m MysqlConfig) (*gorm.DB, error) {
	if m.Dbname == "" {
		return nil, errors.New("db name is empty, please check")
	}
	log.Info("init gorm start: %+v", m)
	var db *gorm.DB
	var err error
	if m.Path != "" {
		db, err = singleMode(m)
		if err != nil {
			return nil, err
		}
	} else {
		db, err = readWriteSplitMode(m)
		if err != nil {
			return nil, err
		}
	}

	log.Info("init gorm gorm.open done ")
	injectMysqlTracing(!m.DisableTrace, db)
	log.Info("init grom inject mysql tracing done ")

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)
	if m.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(m.ConnMaxLifetime) * time.Second)
	}
	log.Info("init gorm all done")
	return db, nil
}

func readWriteSplitMode(m MysqlConfig) (*gorm.DB, error) {
	writePath := strings.Split(m.WritePath, ",")
	m.Path = writePath[0]
	db, err := singleMode(m)
	if err != nil {
		return nil, err
	}

	cfg := dbresolver.Config{
		Sources:           nil,
		Replicas:          nil,
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}
	for _, v := range writePath {
		cfg.Sources = append(cfg.Sources, mysql.New(buildDsn(m.Username, m.Password, v, m.Dbname, m.Config)))
	}
	for _, v := range strings.Split(m.ReadPath, ",") {
		cfg.Replicas = append(cfg.Replicas, mysql.New(buildDsn(m.Username, m.Password, v, m.Dbname, m.Config)))
	}
	log.Info("start to register db resolver %+v", cfg)

	resolver := dbresolver.Register(cfg).SetMaxOpenConns(m.MaxOpenConns).SetMaxIdleConns(m.MaxIdleConns)
	if m.ConnMaxLifetime > 0 {
		resolver.SetConnMaxLifetime(time.Duration(m.ConnMaxLifetime) * time.Second)
	}
	err = db.Use(resolver)
	if err != nil {
		log.Error("gorm init db err %+v", err)
		return nil, err
	}
	return db, nil
}

func singleMode(m MysqlConfig) (*gorm.DB, error) {
	mysqlConfig := buildDsn(m.Username, m.Password, m.Path, m.Dbname, m.Config)
	log.Info("init gorm mysqlConfig: %+v", mysqlConfig)
	db, err := gorm.Open(mysql.New(mysqlConfig))
	if err != nil {
		log.Error("gorm init db err %+v", err)
		return nil, err
	}
	return db, nil
}

func buildDsn(username, password, path, dbname, config string) mysql.Config {
	dsn := username + ":" + password + "@tcp(" + path + ")/" + dbname + "?" + config
	mysqlConfig := mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         191,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}
	return mysqlConfig
}

func injectMysqlTracing(enableTrace bool, db *gorm.DB) {
	if enableTrace {
		if err := db.Use(tracing.NewPlugin(tracing.WithDBSystem(db.Name()))); err != nil {
			log.Error("inject mysql tracing plugin failed with error %v", err)
		} else {
			log.Info("inject mysql tracing plugin")
		}
	}
	return
}

func AddTraceLogger(db *gorm.DB, disableLog bool) *gorm.DB {
	defaultLogger := logger.Default
	if db.Logger != nil {
		defaultLogger = db.Logger
	}
	db.Logger = &traceLogger{Interface: defaultLogger, disableLog: disableLog}
	return db
}

type traceLogger struct {
	logger.Interface
	disableLog bool
}

// Trace implement logger interface
func (l *traceLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}

	if !l.disableLog {
		if err == nil {
			log.CtxInfo(ctx, "[%v][rows:%v] %s", elapsed, rows, sql)
		} else {
			log.CtxWarn(ctx, "[%v][rows:%v] %s, err %+v", elapsed, rows, sql, err)
		}
	} else {
		if err == nil {
			log.CtxDebug(ctx, "[%v][rows:%v] %s", elapsed, rows, sql)
		} else {
			log.CtxError(ctx, "[%v][rows:%v] %s, err %+v", elapsed, rows, sql, err)
		}
	}

	addDbMetrics(mysqlDb, elapsed.Milliseconds(), err)
}
