// Package mongo provide mongo utils
package mongo

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/goriller/ginny-util/graceful"
	codecs "github.com/ti/mongo-go-driver-protobuf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/tag"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

// Mongo the mongo client
type Mongo struct {
	*mongo.Client
	defaultDatabase string
}

var defaultRegistry *bsoncodec.Registry

func init() {
	codec, err := bsoncodec.NewStructCodec(bsoncodec.JSONFallbackStructTagParser)
	if err != nil {
		err = fmt.Errorf("error creating json StructCodec: %w", err)
		panic(err)
	}
	rb := bson.NewRegistryBuilder().
		RegisterDefaultEncoder(reflect.Struct, codec).
		RegisterDefaultDecoder(reflect.Struct, codec)
	defaultRegistry = codecs.Register(rb).Build()
}

// New new client
func NewMongo(ctx context.Context, config *Config) (*Mongo, error) {
	m := &Mongo{}
	mongoClient, err := m.newClient(ctx, config)
	if err != nil {
		return nil, err
	}
	// graceful
	graceful.AddCloser(func(ctx context.Context) error {
		return m.Close(ctx)
	})

	m.Client = mongoClient
	return m, nil
}

// Init init the mongo client
func (m *Mongo) newClient(ctx context.Context, config *Config) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, clientOptions(config))
	if err != nil {
		return nil, err
	}
	//首先ping主节点，主节点若up则无错返回。主节点down时，寻找其他可ping节点，若有1个节点up则无错返回。或直到ServerSelectionTimeout返回错误（driver默认30s）
	//使用ping会降低应用弹性，因为有可能节点是短暂down或正在自动故障转移。所以此处保证集群里有一个节点up 则可启动
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

// TransformDocument transform any object to bson documents
func (m *Mongo) TransformDocument(val interface{}) bsonx.Doc {
	buf := make([]byte, 0, 256)
	b, err := bson.MarshalAppendWithRegistry(defaultRegistry, buf[:0], val)
	if err != nil {
		panic(err)
	}
	dis, err := bsonx.ReadDoc(b)
	if err != nil {
		panic(err)
	}
	return dis
}

// clientOptions
func clientOptions(config *Config) *options.ClientOptions {
	opt := options.Client()

	//集群实例设置
	opt.SetHosts(config.Hosts)
	opt.SetReplicaSet(config.ReplicaSet)

	//读首选项设置
	var readPrefOpts []readpref.Option
	if config.ReadPreference.MaxStaleness > 0 {
		readPrefOpts = append(readPrefOpts, readpref.WithMaxStaleness(
			time.Duration(config.ReadPreference.MaxStaleness)*time.Second))
	}
	if len(config.ReadPreference.Tags) > 0 {
		tags := make([]tag.Tag, 0, len(config.ReadPreference.Tags))
		for name, value := range config.ReadPreference.Tags {
			tags = append(tags, tag.Tag{Name: name, Value: value})
		}
		readPrefOpts = append(readPrefOpts, readpref.WithTagSets(tags))
	}
	switch config.ReadPreference.Mode {
	case int(readpref.PrimaryPreferredMode):
		opt.SetReadPreference(readpref.PrimaryPreferred(readPrefOpts...))
	case int(readpref.SecondaryMode):
		opt.SetReadPreference(readpref.Secondary(readPrefOpts...))
	case int(readpref.SecondaryPreferredMode):
		opt.SetReadPreference(readpref.SecondaryPreferred(readPrefOpts...))
	case int(readpref.NearestMode):
		opt.SetReadPreference(readpref.Nearest(readPrefOpts...))
		//未配置、非法值等 均与PrimaryMode配置一致处理：不显式设置read preference，采用driver默认的Primary模式 无其他首选项配置
	}

	//身份认证设置
	if config.Auth.Mechanism != "" || len(config.Auth.MechanismProperties) > 0 || config.Auth.Source != "" ||
		config.Auth.Username != "" || config.Auth.Password != "" || config.Auth.PasswordSet {
		//无Credential时 需要不调用SetAuth()才可正常连接
		opt.SetAuth(options.Credential{
			AuthMechanism:           config.Auth.Mechanism,
			AuthMechanismProperties: config.Auth.MechanismProperties,
			AuthSource:              config.Auth.Source,
			Username:                config.Auth.Username,
			Password:                config.Auth.Password,
			PasswordSet:             config.Auth.PasswordSet,
		})
	}

	//连接与连接池设置
	opt.SetConnectTimeout(time.Duration(config.ConnectTimeout) * time.Second)
	opt.SetMaxConnIdleTime(time.Duration(config.MaxConnIdleTime) * time.Second)
	opt.SetMaxPoolSize(uint64(config.MaxPoolSize))
	opt.SetMinPoolSize(uint64(config.MinPoolSize))

	return opt
}

// Collection 获取 Collection
func (m *Mongo) Collection(colName string) *mongo.Collection {
	return m.Database(m.defaultDatabase).Collection(colName)
}

// DefaultDatabase 获取默认数据库
func (m *Mongo) DefaultDatabase() *mongo.Database {
	return m.Database(m.defaultDatabase)
}

// Close 关闭
func (m *Mongo) Close(ctx context.Context) error {
	err := m.Client.Disconnect(ctx)
	m.Client = nil
	return err
}
