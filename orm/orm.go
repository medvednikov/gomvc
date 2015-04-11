package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strings"

	"medved/q"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
	"gopkg.in/pg.v2"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
)

var (
	db    *sql.DB
	db2   *pg.DB
	dbmap *gorp.DbMap

	query = "SELECT Id, IsActive, address, views  FROM Room limit 10000"
)

type Room struct {
	Id, Views int
	Isactive  bool
	Address   string
}
type Rooms []*Room

func (users *Rooms) New() interface{} {
	u := &Room{}
	*users = append(*users, u)
	return u
}

func initPq() {
	fmt.Println("INIT PQ")
	var err error
	db, err = sql.Open("postgres", "user=alex password=123 dbname=myorm sslmode=disable")
	h(err)
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(Room{}, "room") //.SetKeys(true, "Id")
}

func initPgx() {
	fmt.Println("INIT PGX")
	//conn, err := pgx.Connect(pgx.ConnConfig{
	config := pgx.ConnConfig{
		Host:     "localhost",
		User:     "alex",
		Password: "123",
		Database: "myorm",
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: config})
	h(err)

	db, err = stdlib.OpenFromConnPool(pool)
	h(err)
}

func initGoPg() {
	db2 = pg.Connect(&pg.Options{
		Host:     "localhost",
		User:     "alex",
		Password: "123",
		Database: "myorm",
	})

}

func main() {
	//genData()

	//testpgx()
	//testpq()
	//testgorp()

}

func testpq() {

	//var id, views int
	//var isactive bool
	//var address string

	rows, err := db.Query(query) //.Scan(&id, &address, &isactive, &views)
	h(err)

	//fmt.Println("TIME0=", time.Now().Sub(t0))
	//room := Room{}
	rooms := make([]*Room, 0)
	//

	//t := time.Now()
	//t0 := time.Now()
	err = scanToStructs(rows, &rooms)
	h(err)

	//fmt.Println("TIME pq=", time.Now().Sub(t0))
	//fmt.Printf("RES=%#v\n", rooms[0])
	fmt.Println("PQ RES:", len(rooms), q.Dump(rooms[0]))
}

func testgorp() {
	var rooms2 []*Room
	//t2 := time.Now()
	dbmap.Select(&rooms2, query) //.Scan(&id, &address, &isactive, &views)
	//fmt.Println("TIME gorp=", time.Now().Sub(t2))
	fmt.Println("GORP RES:", len(rooms2), q.Dump(rooms2[0]))
}

func testpgx() {
	rooms := make([]*Room, 0)
	rows, err := db.Query("SELECT id, views, isactive, address FROM Room ")
	//t0 := time.Now()
	err = scanToStructs(rows, &rooms)
	//fmt.Println("TIME pgx=", time.Now().Sub(t0))
	h(err)
	//rows.Columns()
	/*
		for rows.Next() {
			err = rows.Scan(&views, &isactive, &address)
			fmt.Println("AD=", views, isactive, address)
			h(err)
		}
		h(err)
	*/
	fmt.Println("pgx len=", len(rooms), q.Dump(rooms[0]))
}

func testGoPg() {
	var rooms Rooms
	_, err := db2.Query(&rooms, "SELECT id, views, isactive, address FROM Room ")
	h(err)
	fmt.Println("go pg len=", len(rooms), q.Dump(rooms[0]))
}

func scanToStruct(rows *sql.Rows, out interface{}) {
	cols, err := rows.Columns()
	h(err)

	rows.Next()
	fmt.Println("COLS", cols)

	t := reflect.TypeOf(out)
	res := reflect.New(t)
	var dests []interface{}

	for _, col := range cols {
		col := capitalize(col)
		dests = append(dests, res.Elem().FieldByName(col).Addr().Interface())
	}

	err = rows.Scan(dests...)
	h(err)

	out = res.Interface()

	fmt.Println("RES=", q.Dump(out))

	/*
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fmt.Println(field.Name)
		}
	*/

}

type MyRows interface {
	Scan(dest ...interface{}) error
}

func scanToStructs(rows *sql.Rows, out interface{}) error {
	//t0 := time.Now()

	t := reflect.TypeOf(out) // *[]*User
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("orm: A pointer to slice is expected for Select()")
	}

	// Get the slice type referenced by the pointer
	t = t.Elem() // []*User
	if t.Kind() != reflect.Slice {
		return fmt.Errorf("orm: A pointer to slice is expected for Select()")
	}

	// Get the pointer to the struct
	t = t.Elem() // *User

	// Get list of columns
	cols, err := rows.Columns()
	//fmt.Println("COLS!", cols, t.Name(), t.Elem().Kind())
	if err != nil {
		return fmt.Errorf("orm: Columns() call failed: ", err)
	}

	// This is the empty slice we received as the 'out' argument. We will
	// be writing generated objects to it so that 'out' will contain the
	// results after function's execution
	slice := reflect.Indirect(reflect.ValueOf(out))

	//fmt.Println("1", time.Now().Sub(t0))

	dict := map[string]int{
		"id":       0,
		"views":    1,
		"isactive": 2,
		"address":  3,
	}

	// Loop thru all rows and convert them to Go objects
	for rows.Next() {
		res := reflect.New(t.Elem()) // *User
		var dests []interface{}

		for _, col := range cols {
			//col := capitalize(col)
			// rows.Scan requires a list of references to values,
			// so this builds the required list:
			// rows.Scan(&User.Id, &User.Name, &User.Age)

			//dests = append(dests,
			////res.Elem().FieldByName(col).Addr().Interface())

			dests = append(dests,
				res.Elem().Field(dict[col]).Addr().Interface())

		}

		// Scan the values into res
		err = rows.Scan(dests...)
		if err != nil {
			return fmt.Errorf("orm: Scan() failed:", err)
		}

		// Append res to the end result
		slice.Set(reflect.Append(slice, res))
	}

	return nil
}

func test(i interface{}) {
	//*i = 2
}
func genData() {
	for i := 1; i <= 10000; i++ {
		r := &Room{
			Isactive: true,
			Address:  RandomString(20),
			Views:    rand.Intn(500) + 1,
		}
		dbmap.Insert(r)
	}
}

func h(err error) {
	if err != nil {
		log.Fatal("h err: ", err)
	}
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func RandomString(n int) string {
	var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
	p := make([]rune, n)
	for i := range p {
		p[i] = chars[rand.Intn(len(chars))]
	}
	return string(p)
}
