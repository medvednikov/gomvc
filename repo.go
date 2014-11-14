package ezweb

import (
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	"strings"
)

var (
	queryDebug string
)

// Repo is a PostgreSQL repository type
type Repo struct {
	tableName string
	db        *sql.DB

	// dbmap is the gorp mapping object required for mapping rows to objects
	dbmap *gorp.DbMap
}

// Add inserts a new object by running an INSERT statement
// Example:
// user := &User{Name: "Sam", Age: "20"}
// repo.Add(user)
func (r *Repo) Add(u interface{}) error {
	err := r.dbmap.Insert(u)
	h(err) // log the error
	return err
}

// Retrieve searches for a row with a given id and binds the result to a given
// object
func (r *Repo) Retrieve(result interface{}, id int) {
	err := r.SelectOne(result, "id=$1", id)
	h(err)
}

// Update updates a given object by running an UPDATE statement
// Example:
// var user *User
// user = repo.Retrieve(1)
// user.Age++
// repo.Update(user)
func (r *Repo) Update(u interface{}) error {
	_, err := r.dbmap.Update(u)
	h(err)
	return err
}

// SelectOne searches for a row using a given query and binds the result to a
// given object
// Example:
// var user *User
// repo.SelectOne(&user, "Name=$1 AND Email=$2", name, email)
func (r *Repo) SelectOne(res interface{}, qry string, args ...interface{}) error {
	err := r.dbmap.SelectOne(res, r.selectWhere(qry), args...)
	h(err)
	return err
}

// Select executes a search using a given query and binds the result to a given
// slice
// Example:
// var users *[]User
// repo.Select(&users, "Age > 18")
func (r *Repo) Select(res interface{}, query string, args ...interface{}) error {
	_, err := r.dbmap.Select(res, r.selectWhere(query), args...)
	h(err)
	return err
}

// Count executes a given query and returns the number of corresponding rows
func (r *Repo) Count(query string) int {
	if Debug {
		queryDebug = r.q("SELECT COUNT(*) FROM [t] WHERE " + query)
	}
	count, _ := r.dbmap.SelectInt(
		r.q("SELECT COUNT(*) FROM [t] WHERE " + query))
	return int(count)
}

// Init initializes the repository object. It needs a database pointer, name of
// the table, and the type to bind to.
// Example:
// repo.Init(DB, "Users", User{})
func (r *Repo) Init(db *sql.DB, colName string, obj interface{}) {
	r.tableName = colName
	r.db = db
	r.dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	r.dbmap.AddTableWithName(obj, colName).SetKeys(true, "Id")
}

// q is a helper method that performs all necessary operations on a query string
func (r *Repo) q(query string) string {
	return strings.Replace(query, "[t]", r.tableName, -1)
}

// SELECT is a helper method that builds a SELECT * FROM query
func (r *Repo) selectWhere(query string) string {
	if Debug {
		queryDebug = r.q("SELECT * FROM [t] WHERE " + query)
	}
	return r.q("SELECT * FROM [t] WHERE " + query)
}

// h handles errors (logs them)
func h(err error) {
	if err != nil {
		fmt.Println("ezweb sql error:", err)
		fmt.Println("query:", queryDebug, "\n")
	}
}
