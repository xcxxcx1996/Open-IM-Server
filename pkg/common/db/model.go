package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"time"
)

var DB DataBases

type DataBases struct {
	MysqlDB    mysqlDB
	mgoSession *mgo.Session
	redisPool  *redis.Pool
}

func key(dbAddress, dbName string) string {
	return dbAddress + "_" + dbName
}

func init() {
	var mgoSession *mgo.Session
	var err1 error
	//mysql init
	initMysqlDB()
	// mongo init
	mgoDailInfo := &mgo.DialInfo{
		Addrs:     config.Config.Mongo.DBAddress,
		Direct:    config.Config.Mongo.DBDirect,
		Timeout:   time.Second * time.Duration(config.Config.Mongo.DBTimeout),
		Database:  config.Config.Mongo.DBDatabase,
		Source:    config.Config.Mongo.DBSource,
		Username:  config.Config.Mongo.DBUserName,
		Password:  config.Config.Mongo.DBPassword,
		PoolLimit: config.Config.Mongo.DBMaxPoolSize,
	}
	mgoSession, err := mgo.DialWithInfo(mgoDailInfo)
	if err != nil {
		log.NewError("mgo init err", err.Error(), mgoDailInfo)
	}
	if err != nil {
		time.Sleep(time.Duration(30) * time.Second)
		mgoSession, err1 = mgo.DialWithInfo(mgoDailInfo)
		if err1 != nil {
			panic(err1.Error())
		}
	}
	DB.mgoSession = mgoSession
	DB.mgoSession.SetMode(mgo.Monotonic, true)
	c := DB.mgoSession.DB(config.Config.Mongo.DBDatabase).C(cChat)
	err = c.EnsureIndexKey("uid")
	if err != nil {
		panic(err.Error())
	}

	// redis pool init
	DB.redisPool = &redis.Pool{
		MaxIdle:     config.Config.Redis.DBMaxIdle,
		MaxActive:   config.Config.Redis.DBMaxActive,
		IdleTimeout: time.Duration(config.Config.Redis.DBIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				config.Config.Redis.DBAddress,
				redis.DialReadTimeout(time.Duration(1000)*time.Millisecond),
				redis.DialWriteTimeout(time.Duration(1000)*time.Millisecond),
				redis.DialConnectTimeout(time.Duration(1000)*time.Millisecond),
				redis.DialDatabase(0),
				redis.DialPassword(config.Config.Redis.DBPassWord),
			)
		},
	}
}