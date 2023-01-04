package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"webapp/pkg/data"
	"webapp/pkg/repository"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbName   = "users_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB
var testRepo repository.DatabaseRepo

func TestMain(m *testing.M) {
	//	connect to docker; fail if docker is not running
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker; is it running? %s", err)
	}
	pool = p

	//	set up out docker options, specifying the image and so forth
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	//	get a resource (docker image)
	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not start resource: %s", err)
	}

	//	start the image and wait until it is ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Printf("Error connecting to DB: %s\n", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not connect to DB: %s\n", err)
	}

	//	populate the database with empty tables
	err = createTables()
	if err != nil {
		log.Fatalf("Error creating tables: %s", err)
	}
	testRepo = &PostgresDbRepo{DB: testDB}

	//	run tests
	code := m.Run()

	//	clean up
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables() error {
	tableSQL, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = testDB.Exec(string(tableSQL))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func Test_pingDB(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Errorf("Error pinging DB: %s", err)
	}
}

func TestPostgresDBRepoInsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUser(testUser)
	if err != nil {
		t.Errorf("InsertUser returned an error: %s", err)
	}
	if id != 1 {
		t.Errorf("InsertUser returned bad id; expected 1, but got %d", id)
	}
}

func TestPostgresDBRepoAllUsers(t *testing.T) {
	users, err := testRepo.AllUsers()
	if err != nil {
		t.Errorf("AllUsers returned an error: %s", err)
	}
	if len(users) != 1 {
		t.Errorf("AllUsers returned list of wrong length; expected 1, but got %d", len(users))
	}

	testUser := data.User{
		FirstName: "John",
		LastName:  "Smith",
		Email:     "John.Smith@example.com",
		Password:  "jspwd",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, _ = testRepo.InsertUser(testUser)

	users, err = testRepo.AllUsers()
	if err != nil {
		t.Errorf("AllUsers returned an error: %s", err)
	}
	if len(users) != 2 {
		t.Errorf("AllUsers returned list of wrong length after second insert; expected 2, but got %d", len(users))
	}

}

func TestPostgresDBRepoGetUser(t *testing.T) {
	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Errorf("GetUser returned an error: %q", err)
	} else if user.Email != "admin@example.com" {
		t.Errorf("GetUser returned user with wrong email; expected \"admin@example.com\", but got %q", user.Email)
	}

	_, err = testRepo.GetUser(5)
	if err == nil {
		t.Errorf("No error reported by GetUser when trying to get non-existent user by id")
	}
}

func TestPostgresDBRepoGetUserByEmail(t *testing.T) {
	user, err := testRepo.GetUserByEmail("John.Smith@example.com")
	if err != nil {
		t.Errorf("GetUserByEmail returned an error: %q", err)
	} else if user.ID != 2 {
		t.Errorf("GetUserByEmail returned user with wrong id; expected 2, but got %d", user.ID)
	}
}

func TestPostgresDBRepoUpdateUser(t *testing.T) {
	user, _ := testRepo.GetUser(2)
	user.FirstName = "Jane"
	user.Email = "Jane.Smith@example.com"

	err := testRepo.UpdateUser(*user)
	if err != nil {
		t.Errorf("Error updating user %d: %q", user.ID, err)
	}

	user, _ = testRepo.GetUser(2)
	if user.FirstName != "Jane" || user.Email != "Jane.Smith@example.com" {
		t.Errorf("User 2 has not acually been updated. Expected FirstName %q, but got %q. Expected email %q, but got %q.", "Jane", user.FirstName, "Jane.Smith@example.com", user.Email)
	}
}

func TestPostgresDBRepoDeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(2)
	if err != nil {
		t.Errorf("DeleteUser returned an error: %q", err)
		return
	}
	_, err = testRepo.GetUser(2)
	if err == nil {
		t.Error("Error testing DeleteUser. User ID = 2 was not deleted")
	}
}

func TestPostgresDBRepoResetPassword(t *testing.T) {
	err := testRepo.ResetPassword(1, "password")
	if err != nil {
		t.Errorf("Error changing user's password: %q", err)
	}

	user, _ := testRepo.GetUser(1)
	matches, err := user.PasswordMatches("password")
	if err != nil {
		t.Errorf("Error checking user's password: %q", err)
	}

	if !matches {
		t.Error("Password does not match to the one it was changed to!!!")
	}
}

func TestPostgresDBRepoInsertUserImage(t *testing.T) {
	var image = data.UserImage{
		FileName:  "MyFile.jpg",
		UserID:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	newID, err := testRepo.InsertUserImage(image)
	if err != nil {
		t.Errorf("Error inserting user's image: %q", err)
	}
	if newID != 1 {
		t.Errorf("ID of the first inserted user's image is not equal to 1, but equal to %d", newID)
	}
	image.UserID = 100
	_, err = testRepo.InsertUserImage(image)
	if err == nil {
		t.Errorf("Inserted image for non-existent user!")
	}
}
