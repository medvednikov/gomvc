package repo

import (
	"database/sql"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
)

var (
	queryDebug string
	db         *sql.DB
	Dbmap      *gorp.DbMap

	typeMap = make(map[string]string, 0)
)

type M map[string]interface{}

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
func Add(u interface{}) error {
	err := Dbmap.Insert(u)
	h(err) // log the error
	return err
}

// Retrieve searches for a row with a given id and binds the result to a given
// object
func Retrieve(result interface{}, id int) {
	err := SelectOne(result, "WHERE id=$1", id)
	h(err)
}

// Update updates a given object by running an UPDATE statement
// Example:
// var user *User
// user = repo.Retrieve(1)
// user.Age++
// repo.Update(user)
func Update(u interface{}) error {
	_, err := Dbmap.Update(u)
	h(err)
	return err
}

func Exec(query string, args ...interface{}) error {
	_, err := Dbmap.Exec(query, args...)
	h(err)
	return err
}

// SelectOne searches for a row using a given query and binds the result to a
// given object
// Example:
// var user *User
// repo.SelectOne(&user, "Name=$1 AND Email=$2", name, email)
func SelectOne(res interface{}, query string, args ...interface{}) error {
	var err error
	if strings.Index(query, "SELECT") == -1 {
		query = selectWhere(res, query)
	}

	timeout(func() {
		err = Dbmap.SelectOne(res, query, args...)
		h(err)
	})
	return err
}

// Select executes a search using a given query and binds the result to a given
// slice
// Example:
// var users []*User
// repo.Select(&users, "Age > 18")
func Select(res interface{}, query string, args ...interface{}) error {
	var err error
	if strings.Index(query, "SELECT") == -1 {
		query = selectWhere(res, query)
	}
	timeout(func() {
		_, err = Dbmap.Select(res, query, args...)
		h(err)
	})
	return err
}

func timeout(fn func()) {
	c := make(chan bool, 1)
	go func() {
		fn()
		c <- true
	}()
	select {
	case res := <-c:
		_ = res
	case <-time.After(time.Millisecond * 500):
		log.Println("timeout 1 second")
		fn()
	}
}

// selectWhere is a helper method that builds a SELECT * FROM query
func selectWhere(res interface{}, query string) string {
	//if Debug {
	//queryDebug = r.q("SELECT * FROM [t] WHERE " + query)
	//}
	name := typeName(res)
	return "SELECT * FROM " + typeMap[name] + " " + query
}

func FindAll(res interface{}) error {
	err := Select(res, "")
	h(err)
	return err
}

func SelectInt(query string, args ...interface{}) int64 {
	res, err := Dbmap.SelectInt(query, args...)
	h(err)
	return res
}

// Count executes a given query and returns the number of corresponding rows
/*
func Count(query string) int {
	if Debug {
		queryDebug = r.q("SELECT COUNT(*) FROM [t] WHERE " + query)
	}
	count, _ := Dbmap.SelectInt(
		r.q("SELECT COUNT(*) FROM [t] WHERE " + query))
	return int(count)
}
*/

func typeName(obj interface{}) string {
	name := reflect.TypeOf(obj).String()
	name = name[strings.Index(name, ".")+1:]
	return name
}

// Init initializes the repository object. It needs a database pointer, name of
// the table, and the type to bind to.
// Example:
// repo.Init(DB, "Users", User{})
func InitRepo(dbb *sql.DB, tables M) {
	db = dbb
	Dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	//if gomvc.isDev {
	//Dbmap.TraceOn("", log.New(os.Stdout, "", log.Lmicroseconds))
	//}
	for tableName, obj := range tables {
		typ := typeName(obj)
		typeMap[typ] = tableName
		Dbmap.AddTableWithName(obj, tableName).SetKeys(true, "Id")
	}

}

// h handles errors (logs them)
func h(err error) {
	if err != nil {
		if strings.Index(err.Error(), "no rows in result set") == -1 {
			log.Println("gomvc sql error:", err)
		}
		//fmt.Println("query:", queryDebug, "\n")
		//panic(err)
	}
}
