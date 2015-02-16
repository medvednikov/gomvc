package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"medved/q"
	"reflect"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	_ "github.com/lib/pq"
)

type Room struct {
	Id, Views int
	Isactive  bool
	Address   string
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

func scanToStructs(rows *sql.Rows, out interface{}) error {
	fmt.Println("\n\n")

	t0 := time.Now()

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
	fmt.Println("COLS", cols, t.Name(), t.Elem().Kind())
	if err != nil {
		return fmt.Errorf("orm: Columns() call failed: ", err)
	}

	// This is the empty slice we received as the 'out' argument. We will
	// be writing generated objects to it so that 'out' will contain the
	// results after function's execution
	slice := reflect.Indirect(reflect.ValueOf(out))

	fmt.Println("1", time.Now().Sub(t0))

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
	fmt.Println("\n\n")

	return nil
}

func test(i interface{}) {
	//*i = 2
}

func main() {
	db, err := sql.Open("postgres", "user=alex password=123 dbname=myorm sslmode=disable")
	h(err)
	q := ("SELECT Id, IsActive, address, views  FROM Room limit 10000")

	//var id, views int
	//var isactive bool
	//var address string

	t0 := time.Now()
	rows, err := db.Query(q) //.Scan(&id, &address, &isactive, &views)
	h(err)

	//fmt.Println("TIME0=", time.Now().Sub(t0))
	//room := Room{}
	rooms := make([]*Room, 0)
	//

	//t := time.Now()
	err = scanToStructs(rows, &rooms)
	h(err)
	fmt.Println("TIME=", time.Now().Sub(t0))
	//fmt.Println("RES=", util.Dump(rooms))
	return

	//for _, room := range rooms {
	//fmt.Println(room.Id, room.Isactive, room.Address, room.Views)
	//}

	//fmt.Println(id, address, isactive, views)

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(Room{}, "room") //.SetKeys(true, "Id")

	/*
		for i := 1; i <= 10000; i++ {
			r := &Room{
				Isactive: true,
				Address:  RandomString(20),
				Views:    rand.Intn(500) + 1,
			}
			dbmap.Insert(r)

		}
	*/

	//room2 := Room{}

	var rooms2 []*Room
	t2 := time.Now()
	dbmap.Select(&rooms2, q) //.Scan(&id, &address, &isactive, &views)
	fmt.Println("TIME2=", time.Now().Sub(t2))
	fmt.Println(len(rooms2))
}

func h(err error) {
	if err != nil {
		log.Fatal(err)
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
