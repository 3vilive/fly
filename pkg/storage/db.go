package storage

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/3vilive/fly/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type atomicBool int32

func (b *atomicBool) isSet() bool { return atomic.LoadInt32((*int32)(b)) != 0 }
func (b *atomicBool) setTrue()    { atomic.StoreInt32((*int32)(b), 1) }
func (b *atomicBool) setFalse()   { atomic.StoreInt32((*int32)(b), 0) }

const (
	AdapterMySQL = "mysql"
)

type WrapGormDb struct {
	db *gorm.DB
}

func (w *WrapGormDb) Close() error {
	idb, err := w.db.DB()
	if err != nil {
		return err
	}

	return idb.Close()
}

type DatabaseManager struct {
	config      *viper.Viper
	databaseMap map[string]*WrapGormDb
	mu          sync.Mutex
	closed      atomicBool
}

func NewDatabaseManager(config *viper.Viper) *DatabaseManager {
	return &DatabaseManager{
		databaseMap: make(map[string]*WrapGormDb),
		config:      config,
	}
}

func (m *DatabaseManager) GetDatabase(name string) (*gorm.DB, error) {
	if m.closed.isSet() {
		return nil, errors.New("database manager closed")
	}

	wrapDb, ok := m.databaseMap[name]
	if ok {
		return wrapDb.db, nil
	}

	// init db
	configOfDb := m.config.Sub(name)
	if configOfDb == nil {
		return nil, errors.New("no database config")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// check again, maybe other goroutine init database
	if wrapDb, ok := m.databaseMap[name]; ok {
		return wrapDb.db, nil
	}

	adapater := configOfDb.GetString("adapter")
	var (
		db  *gorm.DB
		err error
	)
	switch adapater {
	case AdapterMySQL:
		db, err = gorm.Open(mysql.Open(configOfDb.GetString("dsn")))
		if err != nil {
			return nil, errors.Wrap(err, "open mysql db conn error")
		}
	default:
		return nil, fmt.Errorf("unsupported adapater: %s", adapater)
	}

	if configOfDb.GetBool("debug") {
		db.Logger.LogMode(logger.Info)
	}
	if idb, err := db.DB(); err != nil {
		idb.SetConnMaxLifetime(2 * time.Hour)

		if maxIdleConns := configOfDb.GetInt("max-idle-connections"); maxIdleConns != 0 {
			idb.SetMaxIdleConns(maxIdleConns)
		}

		if maxOpenConns := configOfDb.GetInt("max-open-connections"); maxOpenConns != 0 {
			idb.SetMaxOpenConns(maxOpenConns)
		}
	}

	wrapDb = &WrapGormDb{
		db: db,
	}
	m.databaseMap[name] = wrapDb

	return wrapDb.db, nil
}

func (m *DatabaseManager) Close() error {
	if m.closed.isSet() {
		return errors.New("database manager closed")
	}
	m.closed.setTrue()

	m.mu.Lock()
	defer m.mu.Unlock()
	for name, wrapDb := range m.databaseMap {
		if err := wrapDb.Close(); err != nil {
			log.Error("close database error", zap.String("database", name), zap.Error(err))
		}

		delete(m.databaseMap, name)
	}

	return nil
}

func GetDatabase(name string) (*gorm.DB, error) {
	if dbManager == nil {
		return nil, errors.New("database manager not init")
	}
	return dbManager.GetDatabase(name)
}
