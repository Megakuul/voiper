package store

import (
	"database/sql"
	"fmt"
)

func migrate(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var version int
	if err = tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version > 3 {
		return fmt.Errorf("database schema %d requires a newer Voiper", version)
	}
	if version == 0 {
		_, err = tx.Exec(`CREATE TABLE contacts (id INTEGER PRIMARY KEY, name TEXT NOT NULL, address TEXT NOT NULL UNIQUE, notes TEXT NOT NULL DEFAULT '', source TEXT NOT NULL DEFAULT 'local', favorite INTEGER NOT NULL DEFAULT 0);
CREATE TABLE history (id INTEGER PRIMARY KEY, account TEXT NOT NULL, remote TEXT NOT NULL, direction TEXT NOT NULL, status TEXT NOT NULL, started TEXT NOT NULL, ended TEXT NOT NULL);
CREATE TABLE messages (id INTEGER PRIMARY KEY, account TEXT NOT NULL, remote TEXT NOT NULL, body TEXT NOT NULL, direction TEXT NOT NULL, status TEXT NOT NULL, created TEXT NOT NULL);`)
		if err != nil {
			return err
		}
	}
	if version < 2 {
		_, err = tx.Exec(`ALTER TABLE contacts ADD COLUMN account TEXT NOT NULL DEFAULT '';
ALTER TABLE contacts ADD COLUMN updated TEXT NOT NULL DEFAULT '';
CREATE TABLE contact_numbers (contact_id INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE, label TEXT NOT NULL, address TEXT NOT NULL, position INTEGER NOT NULL, PRIMARY KEY(contact_id,address));
ALTER TABLE messages ADD COLUMN read INTEGER NOT NULL DEFAULT 1;
CREATE INDEX messages_conversation ON messages(account,remote,id);
CREATE INDEX messages_unread ON messages(read,account);
CREATE INDEX history_account ON history(account,id);
CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);
PRAGMA user_version=2;`)
		if err != nil {
			return err
		}
	}
	if version < 3 {
		_, err = tx.Exec(`CREATE TABLE directory_entries (
remote_id TEXT PRIMARY KEY, contact_id INTEGER REFERENCES contacts(id) ON DELETE SET NULL,
payload TEXT NOT NULL, hidden INTEGER NOT NULL DEFAULT 0);
PRAGMA user_version=3;`)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
