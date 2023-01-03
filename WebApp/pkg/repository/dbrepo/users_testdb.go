package dbrepo

import (
	"database/sql"
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
	var user data.User
	return &user, nil
}

func (m *TestDbRepo) GetUserByEmail(email string) (*data.User, error) {
	var user data.User
	return &user, nil
}

func (m *TestDbRepo) UpdateUser(u data.User) error {
	return nil
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
