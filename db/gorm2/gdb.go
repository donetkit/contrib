package gorm2

import (
	"github.com/donetkit/contrib/utils/db"
	"gorm.io/gorm"
)

type DB struct {
	Client *gorm.DB
}

// Model specify the model you would like to run db operations
//    // update all users is name to `hello`
//    db.Model(&User{}).Update("name", "hello")
//    // if user's primary key is non-blank, will use it as condition, then will only update the user's name to `hello`
//    db.Model(&user).Update("name", "hello")
func (db *DB) Model(value interface{}) db.IDB {
	db.Client = db.Client.Model(value)
	return db
}

// Table specify the table you would like to run db operations
func (db *DB) Table(name string, args ...interface{}) db.IDB {
	db.Client = db.Client.Table(name, args)
	return db
}

// Distinct specify distinct fields that you want querying
func (db *DB) Distinct(args ...interface{}) db.IDB {
	db.Client = db.Client.Distinct(args)
	return db
}

// Select specify fields that you want when querying, creating, updating
func (db *DB) Select(query interface{}, args ...interface{}) db.IDB {
	db.Client = db.Client.Select(query, args)
	return db
}

// Omit specify fields that you want to ignore when creating, updating and querying
func (db *DB) Omit(columns ...string) db.IDB {
	db.Client = db.Client.Omit(columns...)
	return db
}

// Where add conditions
func (db *DB) Where(query interface{}, args ...interface{}) db.IDB {
	db.Client = db.Client.Where(query, args...)
	return db
}

// Not add NOT conditions
func (db *DB) Not(query interface{}, args ...interface{}) db.IDB {
	db.Client = db.Client.Not(query, args)
	return db
}

// Or add OR conditions
func (db *DB) Or(query interface{}, args ...interface{}) db.IDB {
	db.Client = db.Client.Or(query, args...)
	return db
}

// Joins specify Joins conditions
//     db.Joins("Account").Find(&user)
//     db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
//     db.Joins("Account", DB.Select("id").Where("user_id = users.id AND name = ?", "someName").Model(&Account{}))
func (db *DB) Joins(query string, args ...interface{}) db.IDB {
	db.Client = db.Client.Joins(query, args...)
	return db
}

// Group specify the group method on the find
func (db *DB) Group(name string) db.IDB {
	db.Client = db.Client.Group(name)
	return db
}

// Having specify HAVING conditions for GROUP BY
func (db *DB) Having(query interface{}, args ...interface{}) db.IDB {
	db.Client = db.Client.Having(query, args...)
	return db
}

// Order specify order when retrieve records from database
//     db.Order("name DESC")
//     db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func (db *DB) Order(value interface{}) db.IDB {
	db.Client = db.Client.Order(value)
	return db
}

// Limit specify the number of records to be retrieved
func (db *DB) Limit(limit int) db.IDB {
	db.Client = db.Client.Limit(limit)
	return db
}

// Offset specify the number of records to skip before starting to return the records
func (db *DB) Offset(offset int) db.IDB {
	db.Client = db.Client.Offset(offset)
	return db
}

// Scopes pass current database connection to arguments `func(DB) DB`, which could be used to add conditions dynamically
//     func AmountGreaterThan1000(db *gorm.DB) *gorm.DB {
//         return db.Where("amount > ?", 1000)
//     }
//
//     func OrderStatus(status []string) func (db *gorm.DB) *gorm.DB {
//         return func (db *gorm.DB) *gorm.DB {
//             return db.Scopes(AmountGreaterThan1000).Where("status in (?)", status)
//         }
//     }
//
//     db.Scopes(AmountGreaterThan1000, OrderStatus([]string{"paid", "shipped"})).Find(&orders)
//func (db *DB) Scopes(fns ...func(db.IDB) db.IDB) db.IDB {
//	db.Client = db.Client.Scopes(fns...)
//	return db
//}

// Preload associations with given conditions
//    db.Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
func (db *DB) Preload(query string, args ...interface{}) db.IDB {
	db.Client = db.Client.Preload(query, args)
	return db
}

func (db *DB) Attrs(attrs ...interface{}) db.IDB {
	db.Client = db.Client.Attrs(attrs)
	return db
}

func (db *DB) Assign(attrs ...interface{}) db.IDB {
	db.Client = db.Client.Assign(attrs)
	return db
}

func (db *DB) Unscoped() db.IDB {
	db.Client = db.Client.Unscoped()
	return db
}

func (db *DB) Raw(sql string, values ...interface{}) db.IDB {
	db.Client = db.Client.Raw(sql, values)
	return db
}
