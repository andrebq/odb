package odb

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"
)

type (
	// Col represents a collection of objects
	Col struct {
		db *db
		id uuid.UUID
	}
)

func (c *Col) PutObject(key string, content interface{}) (uuid.UUID, error) {
	oid := uuid.NewV5(c.id, key)
	jsonContent, err := json.Marshal(content)
	if err != nil {
		return uuid.UUID{}, err
	}
	_, err = c.db.conn.Exec(`
		insert into fda_dbs.user_objs(id, key, db_id, col_id, content)
		values($1,$2,$3,$4,$5::jsonb)
		on conflict (db_id, id) do update set content = excluded.content;`,
		oid.String(), key, c.db.id, c.id, jsonContent)
	return oid, err
}

func (c *Col) GetObject(out interface{}, key string) error {
	oid := uuid.NewV5(c.id, key)
	var jsonContent string
	err := c.db.conn.QueryRow(`
	select content from fda_dbs.user_objs
	where id = $1 and db_id = $2 and col_id = $3
	`, oid, c.db.id, c.id).Scan(&jsonContent)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(jsonContent), &out)
}
