package repository

import "gorm.io/gorm"

type Database interface {
	Scan(dest interface{}) Database
	Raw(sql string, values ...interface{}) Database
	Error() error
}

type GormDatabase struct {
	DB *gorm.DB
}

func (g *GormDatabase) Scan(dest interface{}) Database {
	g.DB = g.DB.Scan(dest)
	return g
}

func (g *GormDatabase) Raw(sql string, values ...interface{}) Database {
	g.DB = g.DB.Raw(sql, values...)
	return g
}

func (g *GormDatabase) Error() error {
	return g.DB.Error
}
