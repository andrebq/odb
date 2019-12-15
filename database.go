package odb

import (
	"database/sql"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type (
	db struct {
		conn    *sql.DB
		name    string
		id      uuid.UUID
		ownerID uuid.UUID
	}
)

func (d *db) createCollection(name string) (*Col, error) {
	colid := uuid.NewV5(d.id, name)
	err := d.conn.QueryRow(`
	insert into fda_dbs.user_cols(id, db_id, name)
	values ($1, $2, $3)
	returning id`, colid, d.id, name).Scan(&colid)
	if err != nil {
		return nil, err
	}

	return &Col{id: colid, db: d}, nil
}

func (d *db) logentry() *logrus.Entry {
	return logrus.WithField("subsys", "couchdb-server").
		WithField("db", d.id).
		WithField("ownerID", d.ownerID)
}
