package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Database struct {
	gormDb *gorm.DB

	url      string
	username string
	password string
	name     string
}

func NewDatabase(url, user, pass, name string) (*Database, error) {
	db := &Database{
		url:      url,
		username: user,
		password: pass,
		name:     name,
	}

	err := db.init()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *Database) init() error {
	cfg := mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", db.username, db.password, db.url, 3306, db.name),
		// DisableWithReturning: true,
	}
	dialector := mysql.New(cfg)
	logger := gormLogger.New(log.New(os.Stdout, "", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      false,
		Colorful:                  true,
	})
	gormDb, err := gorm.Open(dialector, &gorm.Config{Logger: logger})
	if err != nil {
		return err
	}

	err = gormDb.AutoMigrate(&ServerEntry{})
	if err != nil {
		return err
	}

	db.gormDb = gormDb
	return nil
}

func (db *Database) Write(ip string, port uint16, info *mcpinger.ServerInfo) error {
	entry := &ServerEntry{}
	entry.FromServerInfo(ip, port, info)
	return db.gormDb.Save(entry).Error
}

type ServerEntry struct {
	Ip          string    `gorm:"primaryKey"`
	Port        uint16    `gorm:"primaryKey"`
	LastChecked time.Time `gorm:"autoUpdateTime:true"`

	VersionName     string
	VersionProtocol int32
	Slots           int32
	Online          int32

	Motd string
}

func (se *ServerEntry) FromServerInfo(ip string, port uint16, info *mcpinger.ServerInfo) {
	se.Ip = ip
	se.Port = port
	se.LastChecked = time.Now()
	se.VersionName = info.Version.Name
	se.VersionProtocol = info.Version.Protocol
	se.Slots = info.Players.Max
	se.Online = info.Players.Online
	se.Motd = strings.Trim(info.Description.Text, " \n")
}
