package sqlite3_test

import (
	"database/sql"
	"fmt"
	"os"

	sqlite3 "github.com/coinstack/go-sqlite3"
)

const (
	driver   = "litereplica"
	pitrPath = "binlogs"
	dataSrc  = "file:" + pitrPath + "/foo.db?pitr=on&single_connection=true"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func ExamplePitr() {
	_ = os.RemoveAll(pitrPath)
	err := os.MkdirAll(pitrPath, 0700)
	checkErr(err)

	var sqlite3conn *sqlite3.SQLiteConn
	sqlite3connList := []*sqlite3.SQLiteConn{}
	sql.Register(driver, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			sqlite3conn = conn
			sqlite3connList = append(sqlite3connList, conn)
			return nil
		},
	})
	db, err := sql.Open(driver, dataSrc)
	checkErr(err)

	err = db.Ping()
	checkErr(err)

	point, err := sqlite3conn.GetLastRecoveryPoint()
	checkErr(err)
	fmt.Printf("FirstPoint[%s]\n", point)
	point, err = sqlite3conn.GetLastRecoveryPoint()
	checkErr(err)
	fmt.Printf("LastPoint[%s]\n", point)

	_, err = db.Exec("create table foo (id integer not null primary key, name text)")
	checkErr(err)

	tx, _ := db.Begin()
	_, err = tx.Exec("insert into foo(name) values ('ktlee'), ('sjwoo')")
	checkErr(err)
	tx.Commit()

	rp, err := sqlite3conn.GetLastRecoveryPoint()
	checkErr(err)
	fmt.Printf("[%s]\n", rp)

	tx, _ = db.Begin()
	_, err = tx.Exec("insert into foo(name) values ('kslee'), ('hypark')")
	checkErr(err)
	tx.Commit()

	tx, _ = db.Begin()
	_, err = tx.Exec("update foo set name = 'mlogue' where id = 1")
	checkErr(err)
	tx.Commit()

	point, err = sqlite3conn.GetLastRecoveryPoint()
	checkErr(err)
	fmt.Printf("[%s]\n", point)

	var name string
	rows, err := db.Query("select name from foo order by id")
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&name)
		checkErr(err)
		fmt.Println(name)
	}
	err = rows.Err()
	checkErr(err)

	err = rows.Close()
	checkErr(err)

	count, err := sqlite3conn.RestoreToPoint(rp)
	checkErr(err)
	fmt.Printf("restore to %s, count: %d\n", rp, count)

	row := db.QueryRow("select name from foo where id = 4")
	err = row.Scan(&name)
	checkErr(err)
	if err == nil {
		fmt.Println(name)
	}

	row = db.QueryRow("select name from foo where id = 1")
	err = row.Scan(&name)
	checkErr(err)
	if err == nil {
		fmt.Println(name)
	}

	err = db.Close()
	checkErr(err)

	fmt.Println("opening new db")
	db, err = sql.Open(driver, dataSrc)
	checkErr(err)

	err = db.Ping()
	checkErr(err)

	tx, _ = db.Begin()
	_, err = tx.Exec("insert into foo(name) values ('kslee'), ('hypark')")
	checkErr(err)
	tx.Commit()

	point, err = sqlite3conn.GetFirstRecoveryPoint()
	fmt.Printf("FirstPoint[%s]\n", point)
	point, err = sqlite3conn.GetLastRecoveryPoint()
	fmt.Printf("LastPoint[%s]\n", point)

	err = db.Close()
	checkErr(err)

	// Output:
	// FirstPoint[]
	// LastPoint[]
	// [foo.db.0000000000000002]
	// [foo.db.0000000000000004]
	// mlogue
	// sjwoo
	// kslee
	// hypark
	// restore to foo.db.0000000000000002, count: 2
	// sql: no rows in result set
	// ktlee
	// opening new db
	// FirstPoint[foo.db.0000000000000001]
	// LastPoint[foo.db.0000000000000003]
}
