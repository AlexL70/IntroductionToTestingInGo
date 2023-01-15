package dbrepo

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"webapp/pkg/data"
)

type TestDbRepo struct{}

func (m *TestDbRepo) Connection() *sql.DB {
	return nil
}

func (m *TestDbRepo) AllUsers() ([]*data.User, error) {
	var users []*data.User
	return users, nil
}

func (m *TestDbRepo) GetUser(id int) (*data.User, error) {
	var user = data.User{}
	if id == 1 {
		user = data.User{
			ID:        1,
			FirstName: "Admin",
			LastName:  "User",
			Email:     "admin@example.com",
		}
		return &user, nil
	}
	return nil, errors.New("user not found")
}

func (m *TestDbRepo) GetUserByEmail(email string) (*data.User, error) {
	if email == "admin@example.com" {
		return &data.User{
			ID:        1,
			FirstName: "Admin",
			LastName:  "User",
			Email:     "admin@example.com",
			Password:  "$2a$14$ajq8Q7fbtFRQvXpdCq7Jcuy.Rx1h/L4J60Otx.gyNLbAYctGMJ9tK",
			IsAdmin:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	} else {
		return nil, fmt.Errorf("user with email %q not found", email)
	}
}

func (m *TestDbRepo) UpdateUser(u data.User) error {
	if u.ID == 1 {
		return nil
	}
	return errors.New("error updating user")
}

func (m *TestDbRepo) DeleteUser(id int) error {
	return nil
}

func (m *TestDbRepo) InsertUser(user data.User) (int, error) {
	return 2, nil
}

func (m *TestDbRepo) ResetPassword(id int, password string) error {
	return nil
}

func (m *TestDbRepo) InsertUserImage(i data.UserImage) (int, error) {
	return 1, nil
}
