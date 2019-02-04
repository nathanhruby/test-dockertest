package dockertest_test

import (
    "testing"
    "log"
    "github.com/ory/dockertest"
    _ "github.com/go-sql-driver/mysql"
    "database/sql"
    "fmt"
    "os"
)

var db *sql.DB

func TestMain(m *testing.M) {
    // uses a sensible default on windows (tcp/http) and linux/osx (socket)
    pool, err := dockertest.NewPool("")
    if err != nil {
        log.Fatalf("Could not connect to docker: %s", err)
    }

    // pulls an image, creates a container based on it and runs it
    resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
    if err != nil {
        log.Fatalf("Could not start resource: %s", err)
    }

    // exponential backoff-retry, because the application in the container might not be ready to accept connections yet
    if err := pool.Retry(func() error {
        var err error
        db, err = sql.Open("mysql", fmt.Sprintf("root:secret@(%s)/mysql", resource.GetHostPort("3306/tcp")))
        if err != nil {
            return err
        }
        return db.Ping()
    }); err != nil {
        log.Fatalf("Could not connect to mysql: %s", err)
    }

    code := m.Run()

    // You can't defer this because os.Exit doesn't care for defer
    if err := pool.Purge(resource); err != nil {
        log.Fatalf("Could not purge resource: %s", err)
    }

    os.Exit(code)
}

func TestSomething(t *testing.T) {
    _, err := db.Query("SELECT 1")
    if err != nil {
        t.Fail()
    }
}

