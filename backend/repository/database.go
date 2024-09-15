package repository

import "gorm.io/gorm"

// we set up a database interface which is will allow us to switch databases in the future
// the interface method also is defined to limit the SQL methods
// this is to prevent unintended database manipulation to be added to the application
type Database interface {
	Model(value interface{}) Database
	Find(dest interface{}, conds ...interface{}) Database
	First(dest interface{}, conds ...interface{}) Database
	Preload(query string, args ...interface{}) Database
	Where(query interface{}, args ...interface{}) Database
	Count(count *int64) Database
	Error() error
}

type GormDatabase struct {
	DB *gorm.DB
}

func (g *GormDatabase) Model(value interface{}) Database {
	g.DB = g.DB.Model(value)
	return g
}

func (g *GormDatabase) Find(dest interface{}, conds ...interface{}) Database {
	g.DB = g.DB.Find(dest, conds...)
	return g
}

func (g *GormDatabase) First(dest interface{}, conds ...interface{}) Database {
	g.DB = g.DB.First(dest, conds...)
	return g
}

func (g *GormDatabase) Preload(query string, args ...interface{}) Database {
	g.DB = g.DB.Preload(query, args...)
	return g
}

func (g *GormDatabase) Where(query interface{}, args ...interface{}) Database {
	g.DB = g.DB.Where(query, args...)
	return g
}

func (g *GormDatabase) Count(count *int64) Database {
	g.DB = g.DB.Count(count)
	return g
}
func (g *GormDatabase) Error() error {
	return g.DB.Error
}
