package main

import "database/sql"

type User struct {
	Token      string
	Username   sql.NullString
	LastIp     sql.NullString
	Limitation int
	CreatedAt  sql.NullTime
	UpdatedAt  sql.NullTime
}
