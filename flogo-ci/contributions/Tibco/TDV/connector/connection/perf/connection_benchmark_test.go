package perf

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

// var conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", "flogo", "adminadmin", "192.168.1.10", 3306, "rakshit")
var conninfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "username", "adminadmin", "arakshit")

var preparedQueryCache map[string]*sql.Stmt
var preparedQueryCacheMutex sync.Mutex

// func main() {
// 	log.Infof("StartTime for  %v\n", time.Now())
// 	BenchmarkMaxOpenConns1()
// 	log.Infof("Stoptime for %v\n", time.Now())
// }

func insertRecord(b *testing.B, db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// fmt.Printf("before exec stat : %v \n", db.Stats())

	query := "insert into perftest values('rakshit','ashtekar')"
	stmt, err := getPreparedStatement(query, db)

	//log.Print("Preparing prepared statement for query...")
	//stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var args interface{}

	_, err = stmt.ExecContext(ctx, args)

	if err != nil {
		stmt.Close()
	}
}

func getPreparedStatement(query string, db *sql.DB) (stmt *sql.Stmt, err error) {
	preparedQueryCacheMutex.Lock()
	defer preparedQueryCacheMutex.Unlock()
	stmt, ok := preparedQueryCache[query]
	if !ok {
		stmt, err = db.Prepare(query)
		if err != nil {
			return nil, err
		}
	}
	return stmt, nil

}

// BenchmarkMaxOpenConns1 func
func BenchmarkMaxOpenConns1(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
	//log.Printf("Connetion string: " + conninfo)
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

	// log.Print("Max Open Connections: " + strconv.Itoa(db.Stats().MaxOpenConnections))
	// log.Print("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	// log.Print("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	// log.Print("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	// log.Print("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	// log.Print("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))

}

//BenchmarkMaxOpenConns2 func
func BenchmarkMaxOpenConns2(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

//BenchmarkMaxOpenConns5 func
func BenchmarkMaxOpenConns5(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

//BenchmarkMaxOpenConns10 func
func BenchmarkMaxOpenConns10(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxOpenConnsUnlimited func
func BenchmarkMaxOpenConnsUnlimited(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxIdleConnsNone func
func BenchmarkMaxIdleConnsNone(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxIdleConns1 func
func BenchmarkMaxIdleConns1(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxIdleConns2 func
func BenchmarkMaxIdleConns2(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxIdleConns5 func
func BenchmarkMaxIdleConns5(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxIdleConns10 func
func BenchmarkMaxIdleConns10(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkConnMaxLifetimeUnlimited func
func BenchmarkConnMaxLifetimeUnlimited(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkConnMaxLifetime1000 func
func BenchmarkConnMaxLifetime1000(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkConnMaxLifetime500 func
func BenchmarkConnMaxLifetime500(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkConnMaxLifetime200 func
func BenchmarkConnMaxLifetime200(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkConnMaxLifetime100 func
func BenchmarkConnMaxLifetime100(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxOpenIdleLifeCombined551 func
func BenchmarkMaxOpenIdleLifeCombined551(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxOpenIdleLifeCombined10102 func
func BenchmarkMaxOpenIdleLifeCombined10102(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxOpenIdleLifeCombined25255 func
func BenchmarkMaxOpenIdleLifeCombined25255(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
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

// BenchmarkMaxOpenIdleLifeCombined25251sec func
func BenchmarkMaxOpenIdleLifeCombined25251sec(b *testing.B) {
	db, err := sql.Open("odbc", conninfo)
	if err != nil {
		b.Fatal(err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	duration, err := time.ParseDuration("2h5m")
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(duration)
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			insertRecord(b, db)
		}
	})
	log.Print("Max Open Connections: " + strconv.Itoa(db.Stats().MaxOpenConnections))
	log.Print("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	log.Print("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	log.Print("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	log.Print("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	log.Print("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))
}
