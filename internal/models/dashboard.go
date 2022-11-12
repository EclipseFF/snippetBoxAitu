package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type UsersTable struct {
	DB *pgxpool.Pool
}

type AdminsTable struct {
	DB *pgxpool.Pool
}

type Users struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type Admins struct {
	ID             int
	Email          string
	HashedPassword []byte
	Created        time.Time
}

func (m *UsersTable) GetUsers() ([]*Users, error) {
	stmt := `SELECT id, name, email, created FROM users`

	rows, err := m.DB.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	showUsers := []*Users{}

	for rows.Next() {

		s := &Users{}

		err = rows.Scan(&s.ID, &s.Name, &s.Email, &s.Created)
		if err != nil {
			return nil, err
		}
		showUsers = append(showUsers, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return showUsers, nil
}

func (m *AdminsTable) InsertAdmin(email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	fmt.Println(email)
	fmt.Println(hashedPassword)

	stmt := `INSERT INTO admins (email, hashed_password, created)
VALUES($1, $2, now())`

	_, err = m.DB.Exec(context.Background(), stmt, email, string(hashedPassword))
	if err != nil {
		var myPGError *pgconn.PgError

		if errors.As(err, &myPGError) {
			if myPGError.Code == "23505" && strings.Contains(myPGError.Error(), "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		return err
	}
	return nil
}

func (m *AdminsTable) AdminAuthenticate(email, password string) (int, error) {

	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM admins WHERE email = $1`
	err := m.DB.QueryRow(context.Background(), stmt, email).Scan(&id, &hashedPassword)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

//func (m *UsersTable) UpdateUsers() ([]*Users, error) {
//	stmt := `SELECT id, name, email, created FROM users`
//	rows, err := m.DB.Query(stmt)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//	showUsers := []*Users{}
//	for rows.Next() {
//		s := &Users{}
//		err = rows.Scan(&s.ID, &s.Name, &s.Email, &s.Created)
//		if err != nil {
//			return nil, err
//		}
//		showUsers = append(showUsers, s)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, err
//	}
//	return showUsers, nil
//}

func (m *UsersTable) RemoveUser(id int) error {
	stmt := `DELETE FROM users where id = $1`
	_, err := m.DB.Exec(context.Background(), stmt, id)
	if err != nil {
		return err
	}
	return nil
}
