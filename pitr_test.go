package sqlite3_test

import (
	"database/sql"
	"fmt"
	"os"

	sqlite3 "bitbucket.org/cloudwallet/go-sqlite3"
)

const (
	driver   = "litereplica"
	pitrPath = "binlogs"
	dataSrc  = "file:./foo.db?pitr=on&pitr_path=" + pitrPath
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func ExamplePitr() {
	_ = os.Remove("./foo.db")
	_ = os.RemoveAll(pitrPath)

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

	fmt.Printf("[%s]\n", sqlite3conn.GetLastRecoveryPoint())

	_, err = db.Exec("create table foo (id integer not null primary key, name text)")
	checkErr(err)

	tx, _ := db.Begin()
	_, err = tx.Exec("insert into foo(name) values ('ktlee'), ('sjwoo')")
	checkErr(err)
	tx.Commit()

	point := sqlite3conn.GetLastRecoveryPoint()
	fmt.Printf("[%s]\n", point)

	tx, _ = db.Begin()
	_, err = tx.Exec("insert into foo(name) values ('kslee'), ('hypark')")
	checkErr(err)
	tx.Commit()

	tx, _ = db.Begin()
	_, err = tx.Exec("update foo set name = 'mlogue' where id = 1")
	checkErr(err)
	tx.Commit()

	fmt.Printf("[%s]\n", sqlite3conn.GetLastRecoveryPoint())

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

	count := sqlite3conn.RestoreToPoint(point)
	fmt.Printf("restore to %s, count: %d\n", point, count)

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

	fmt.Printf("[%s]\n", sqlite3conn.GetLastRecoveryPoint())

	err = db.Close()
	checkErr(err)

	// Output:
	// []
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
	// [foo.db.0000000000000003]
}
