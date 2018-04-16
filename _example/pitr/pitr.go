package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	sqlite3 "bitbucket.org/cloudwallet/go-sqlite3"
)

const (
	driver   = "litereplica"
	pitrPath = "binlogs"
	dataSrc  = "file:./foo.db?pitr=on&pitr_path=" + pitrPath
)

func main() {
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
	if err != nil {
		fmt.Println(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("[%s]\n", sqlite3conn.GetLastRecoveryPoint())

	_, err = db.Exec("create table foo (id integer not null primary key, name text)")
	if err != nil {
		fmt.Println(err)
	}

	tx, _ := db.Begin()
	tx.Exec("insert into foo(name) values ('ktlee'), ('sjwoo')")
	tx.Commit()

	point := sqlite3conn.GetLastRecoveryPoint()
	fmt.Printf("[%s]\n", point)

	tx, _ = db.Begin()
	tx.Exec("insert into foo(name) values ('kslee'), ('hypark')")
	tx.Commit()

	tx, _ = db.Begin()
	//tx.Exec("insert into foo(name) values ('hypark')")
	tx.Exec("update foo set name = 'mlogue' where id = 1")
	tx.Commit()

	fmt.Printf("[%s]\n", sqlite3conn.GetLastRecoveryPoint())

	var name string
	rows, err := db.Query("select name from foo order by id")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(name)
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
	}
	rows.Close()

	count := sqlite3conn.RestoreToPoint(point)
	fmt.Printf("restore to %s, count: %d\n", point, count)

	row := db.QueryRow("select name from foo where id = 4")
	err = row.Scan(&name)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(name)
	}

	row = db.QueryRow("select name from foo where id = 1")
	err = row.Scan(&name)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(name)
	}

	db.Close()

	fmt.Println("opening new db")
	db, err = sql.Open(driver, dataSrc)
	if err != nil {
		fmt.Println(err)
	}

	db.Ping()

	tx, _ = db.Begin()
	tx.Exec("insert into foo(name) values ('kslee'), ('hypark')")
	tx.Commit()

	fmt.Printf("[%s]\n", sqlite3conn.GetLastRecoveryPoint())

	err = db.Close()
	if err != nil {
		fmt.Println(err)
	}
}
