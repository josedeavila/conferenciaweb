package modelo

import (
	"database/sql"
	r "github.com/dancannon/gorethink"
)

var (
	db   *sql.DB
    session *r.Session
)

func Db(d *sql.DB,session *r.Session) {
	db = d
	session = session
}

func GetSession() *r.Session {
    if session == nil {
       // init()
    }
    return session
}