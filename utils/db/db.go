package db

type IDB interface {

	// Model specify the model you would like to run db operations
	//    // update all users is name to `hello`
	//    db.Model(&User{}).Update("name", "hello")
	//    // if user's primary key is non-blank, will use it as condition, then will only update the user's name to `hello`
	//    db.Model(&user).Update("name", "hello")
	Model(value interface{}) IDB

	// Table specify the table you would like to run db operations
	Table(name string, args ...interface{}) IDB

	// Distinct specify distinct fields that you want querying
	Distinct(args ...interface{}) IDB

	// Select specify fields that you want when querying, creating, updating
	Select(query interface{}, args ...interface{}) IDB

	// Omit specify fields that you want to ignore when creating, updating and querying
	Omit(columns ...string) IDB

	// Where add conditions
	Where(query interface{}, args ...interface{}) IDB

	// Not add NOT conditions
	Not(query interface{}, args ...interface{}) IDB

	// Or add OR conditions
	Or(query interface{}, args ...interface{}) IDB

	// Joins specify Joins conditions
	//     db.Joins("Account").Find(&user)
	//     db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").Find(&user)
	//     db.Joins("Account", DB.Select("id").Where("user_id = users.id AND name = ?", "someName").Model(&Account{}))
	Joins(query string, args ...interface{}) IDB

	// Group specify the group method on the find
	Group(name string) IDB
	// Having specify HAVING conditions for GROUP BY
	Having(query interface{}, args ...interface{}) IDB

	// Order specify order when retrieve records from database
	//     db.Order("name DESC")
	//     db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
	Order(value interface{}) IDB

	// Limit specify the number of records to be retrieved
	Limit(limit int) IDB
	// Offset specify the number of records to skip before starting to return the records
	Offset(offset int) IDB

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
	//Scopes(fns ...func(IDB) IDB) IDB

	// Preload associations with given conditions
	//    db.Preload("Orders", "state NOT IN (?)", "cancelled").Find(&users)
	Preload(query string, args ...interface{}) IDB

	Attrs(attrs ...interface{}) IDB

	Assign(attrs ...interface{}) IDB

	Unscoped() IDB

	Raw(sql string, values ...interface{}) IDB
}
