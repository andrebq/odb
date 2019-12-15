package odb

import (
	"database/sql"
	"fmt"
	"strings"

	// only works with pg anyway

	_ "github.com/lib/pq"
	uuid "github.com/satori/go.uuid"

	"github.com/sirupsen/logrus"
)

type (
	// Server represents a connection to a PostgreSQL server
	Server struct {
		id   string
		conn *sql.DB
	}
)

var (
	// serverUUID is the root namespace for all UUID
	serverUUID = uuid.Must(uuid.FromString("453632b9-dcc0-4823-815b-4f36e6ad6353"))
)

// ServerUUID returns the root namespace UUID
func ServerUUID() uuid.UUID {
	return serverUUID
}

// Collection returns a collection of objects for the given user/database pair.
//
// Intermediary objects will be created
func (s *Server) Collection(username, db, col string) (*Col, error) {
	user, err := s.createUser(username)
	if err != nil {
		return nil, err
	}
	dbref, err := s.createDatabase(user, db)
	if err != nil {
		return nil, err
	}
	return dbref.createCollection(col)
}

func Connect(host, user, pwd string) (*Server, error) {
	var port = "5432"
	if idx := strings.Index(host, ":"); idx >= 0 {
		port = host[idx+1:]
		host = host[:idx]
	}
	conn, err := sql.Open("postgres",
		fmt.Sprintf("sslmode=disable user=%v host=%v password=%v dbname=fda_db port=%v", user, host, pwd, port))
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	return &Server{
		id:   host,
		conn: conn,
	}, nil
}

func (s *Server) logentry() *logrus.Entry {
	return logrus.WithField("subsys", "couchdb-server").WithField("id", s.id)
}

func (s *Server) createUser(username string) (uuid.UUID, error) {
	id := uuid.NewV5(ServerUUID(), username)
	role := "owner-" + id.String()
	_, err := s.conn.Exec("insert into fda_users.users(id, name, role) values ($1, $2, $3)",
		id.String(), username, role)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func (s *Server) createDatabase(userid uuid.UUID, database string) (*db, error) {
	dbid := uuid.NewV5(userid, database)
	_, err := s.conn.Exec("insert into fda_dbs.user_dbs(id, owner_id, name) values ($1, $2, $3)",
		dbid.String(), userid, database)
	if err != nil {
		return nil, err
	}
	return &db{
		ownerID: userid,
		id:      dbid,
		name:    database,
		conn:    s.conn,
	}, err
}
