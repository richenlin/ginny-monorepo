package orm

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/wire"
	"github.com/goriller/ginny-gorm/dialector"
	"github.com/goriller/ginny-util/graceful"
	"github.com/goriller/gorm-plus/gplus"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

var Provider = wire.NewSet(NewConfig, New)

// New
func New(ctx context.Context, conf *Config) (*gorm.DB, error) {
	mdb, err := newDB(ctx, conf.Master, conf)
	if err != nil {
		return nil, err
	}
	// RDB多个
	rDBLen := len(conf.Replicas)
	replicas := []gorm.Dialector{}
	for i := 0; i < rDBLen; i++ {
		readDB, err := newDB(ctx, conf.Replicas[i], conf)
		if err != nil {
			return nil, err
		}
		replicas = append(replicas, readDB.Dialector)
	}

	err = mdb.Use(
		dbresolver.Register(dbresolver.Config{
			Sources:  []gorm.Dialector{mdb.Dialector},
			Replicas: replicas,
			// sources/replicas load balancing policy
			Policy: dbresolver.RandomPolicy{},
			// print sources/replicas mode in logger
			TraceResolverMode: true,
		}).SetConnMaxIdleTime(time.Hour).
			SetConnMaxLifetime(24 * time.Hour).
			SetMaxIdleConns(100).
			SetMaxOpenConns(200),
	)
	if err != nil {
		return nil, err
	}

	gplus.Init(mdb)

	return mdb, nil
}

// newDB
func newDB(ctx context.Context, dsn string, conf *Config) (*gorm.DB, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "faild parse dsn")
	}
	switch conf.Type {
	case "sqllite":
		conf.dialector = dialector.NewSqllite(u)
		break
	case "mysql":
		conf.dialector = dialector.NewMysql(u)
		break
	case "postgres":
	case "postgresql":
		conf.dialector = dialector.NewPostgres(u)
		break
	case "sqlserver":
		conf.dialector = dialector.NewSqlserver(u)
		break
	case "clickhouse":
		conf.dialector = dialector.NewClickhouse(u)
		break
	default:
		return nil, errors.Wrap(err, "not support")
	}
	db, err := gorm.Open(conf.dialector, conf)

	// graceful
	graceful.AddCloser(func(ctx context.Context) error {
		dbInstance, err := db.DB()
		if err != nil {
			return err
		}
		return dbInstance.Close()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database")
	}
	return db, nil
}
