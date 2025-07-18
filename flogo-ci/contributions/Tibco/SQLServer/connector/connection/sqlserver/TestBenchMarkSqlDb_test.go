// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"time"

// 	"git.tibco.com/git/product/ipaas/wi-mssql/src/app/github.com/TIBCOSoftware/flogo-lib/logger"
// 	_ "github.com/denisenkom/go-mssqldb"
// )

// var log = logger.GetLogger("flogo.sqlserver_test")

// const (
// 	goroutines = 10
// 	entries    = 10000
// )

// var db *sql.DB

// func main() {
// 	// t0 := time.Unix(1000000, 0)
// 	//
// 	// create string to pass
// 	// var sStmt string = "insert into test values(?,?)"
// 	// conninfo := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;sslmode=disable;",
// 	// 	"192.168.0.105", 1433, "SA", "Sqlserver@9", "SnehalDB")
// 	// db, err := sql.Open("mssql", conninfo)
// 	// if err != nil {
// 	// 	log.Infof("error %v", err)
// 	// }
// 	// db.SetMaxIdleConns(1)
// 	// db.SetMaxOpenConns(6)
// 	// db.SetConnMaxLifetime(5 * time.Second)
// 	// run the insert function using 10 go routines
// 	conninfo := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;sslmode=disable;",
// 		"192.168.0.105", 1433, "SA", "Sqlserver@9", "SnehalDB")
// 	for j := 0; j < goroutines; j++ {
// 		// spin up a go routine
// 		log.Infof("routine :  %d", j)
// 		go UsingGoRoutines(conninfo, j)
// 	}

// 	// for true {

// 	// 	if db != nil {
// 	// 		log.Infof("stats %v", db.Stats())
// 	// 		time.Sleep(5 * time.Second)
// 	// 	}

// 	// }
// 	// this is a simple way to keep a program open
// 	// the go program will close when a key is pressed
// 	var input string
// 	fmt.Scanln(&input)
// }

// // UsingGoRoutines ..
// func UsingGoRoutines(conninfo string, j int) {

// 	// lazily open db (doesn't truly open until first request)
// 	db, err := sql.Open("mssql", conninfo)

// 	if err != nil {
// 		log.Infof("error %v", err)
// 	}

// 	db.SetMaxOpenConns(25)
// 	db.SetMaxIdleConns(25)
// 	db.SetConnMaxLifetime(5 * time.Minute)

// 	// log.Infof("Before stats %v", db.Stats())
// 	stmt, err := db.Prepare("insert into test1 values(?,?)")
// 	if err != nil {
// 		log.Infof("error %v", err)
// 	}
// 	// log.Infof("StartTime: %v\n", time.Now())

// 	log.Infof("StartTime for %d routine: %v\n", j, time.Now())
// 	// log.Infof("Before stats %v", db.Stats())
// 	for i := 0; i < entries; i++ {
// 		// fmt.Printf("for stats %v", db.Stats())
// 		// log.Infof("before stats %v", db.Stats())

// 		res, err := stmt.Exec(i, "name")
// 		if err != nil || res == nil {
// 			log.Infof("error %v", err)
// 		}
// 		// log.Infof("after stats %v", db.Stats())
// 	}
// 	log.Infof("StopTime: for %d routine %v\n", j, time.Now())
// 	stmt.Close()
// 	// close db
// 	db.Close()

// 	// fmt.Printf("StopTime: %v\n", time.Now())

// }

package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var conninfo = fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;sslmode=disable;",
	"192.168.0.105", 1433, "SA", "Sqlserver@9", "SnehalDB")

// func main() {
// 	log.Infof("StartTime for  %v\n", time.Now())
// 	BenchmarkMaxOpenConns1()
// 	log.Infof("Stoptime for %v\n", time.Now())
// }

func insertRecord(b *testing.B, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// fmt.Printf("before exec stat : %v \n", db.Stats())

	_, err := db.ExecContext(ctx, "insert into test1 values(1,'name')")
	// fmt.Printf("after exec stat : %v \n", db.Stats())
	// fmt.Printf("then %v\n", time.Now())
	if err != nil {
		b.Fatal("err : ", err)
	}
}

func BenchmarkMaxOpenConns1(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxOpenConns2(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(2)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxOpenConns5(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(5)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxOpenConns10(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(10)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxOpenConnsUnlimited(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxIdleConnsNone(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxIdleConns(0)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxIdleConns1(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxIdleConns(1)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxIdleConns2(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxIdleConns(2)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxIdleConns5(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxIdleConns(5)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxIdleConns10(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxIdleConns(10)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkConnMaxLifetimeUnlimited(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkConnMaxLifetime1000(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}

	db.SetConnMaxLifetime(1000 * time.Millisecond)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkConnMaxLifetime500(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetConnMaxLifetime(500 * time.Millisecond)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkConnMaxLifetime200(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetConnMaxLifetime(200 * time.Millisecond)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkConnMaxLifetime100(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetConnMaxLifetime(100 * time.Millisecond)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxOpenIdleLifeCombined551(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Minute)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}
func BenchmarkMaxOpenIdleLifeCombined10102(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(2 * time.Minute)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxOpenIdleLifeCombined25255(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}

func BenchmarkMaxOpenIdleLifeCombined25251sec(b *testing.B) {
	db, err := sql.Open("mssql", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(1 * time.Second)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
}
