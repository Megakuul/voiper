package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

// DirectoryProfile contains no secret. Passwords belong to the desktop wallet
// or the running session, never the contact database.
type DirectoryProfile struct {
	URL, BaseDN, BindDN, CAFile string
	Limit, IntervalMinutes      int
	Enabled, UseSecretService   bool
}

type DirectoryStatus struct {
	LastAttempt, LastSuccess time.Time
	Error                    string
	Entries, Preserved       int
	Partial                  bool
}

func (p DirectoryProfile) Validate() error {
	if !p.Enabled && p.URL == "" {
		return nil
	}
	u, err := url.Parse(p.URL)
	if err != nil || (u.Scheme != "ldap" && u.Scheme != "ldaps") || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("use an ldap:// or ldaps:// server URL without credentials, path, or query")
	}
	if len(p.URL) > 2048 || len(p.BaseDN) > 4096 || len(p.BindDN) > 4096 || len(p.CAFile) > 4096 {
		return errors.New("directory configuration is too long")
	}
	if strings.TrimSpace(p.BaseDN) == "" {
		return errors.New("directory base DN is required")
	}
	if _, err := ldap.ParseDN(p.BaseDN); err != nil {
		return errors.New("invalid directory base DN")
	}
	if p.Limit < 1 || p.Limit > 5000 {
		return errors.New("directory limit must be between 1 and 5000")
	}
	if p.IntervalMinutes < 5 || p.IntervalMinutes > 1440 {
		return errors.New("refresh interval must be between 5 and 1440 minutes")
	}
	return nil
}

func (s *Store) DirectoryProfile() (DirectoryProfile, error) {
	p := DirectoryProfile{Limit: 1000, IntervalMinutes: 60}
	var raw string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key='directory-profile'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	err = json.Unmarshal([]byte(raw), &p)
	return p, err
}

func (s *Store) SaveDirectoryProfile(p DirectoryProfile) error {
	if err := p.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var previous string
	err = tx.QueryRow(`SELECT value FROM settings WHERE key='directory-profile'`).Scan(&previous)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var old DirectoryProfile
	if previous != "" {
		if err = json.Unmarshal([]byte(previous), &old); err != nil {
			return err
		}
	}
	if old.URL != p.URL || old.BaseDN != p.BaseDN || old.BindDN != p.BindDN {
		// Changing the directory keeps its old contacts as an imported snapshot.
		// The new server can never delete contacts belonging to the old server.
		if _, err = tx.Exec(`UPDATE contacts SET source='directory' WHERE source='directory-sync'; DELETE FROM directory_entries; DELETE FROM settings WHERE key='directory-status'`); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES('directory-profile',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, string(raw)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) DirectoryStatus() (DirectoryStatus, error) {
	var status DirectoryStatus
	var raw string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key='directory-status'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return status, nil
	}
	if err != nil {
		return status, err
	}
	err = json.Unmarshal([]byte(raw), &status)
	return status, err
}

// ApplyDirectory commits only a result for the still-configured profile. A
// partial scan may update entries but cannot prove that absent entries vanished.
// Edited contacts have source=local, so later refreshes never overwrite them.
func (s *Store) ApplyDirectory(profile DirectoryProfile, result DirectoryResult, lookupErr error) error {
	if len(result.Contacts) > 5000 {
		return errors.New("directory result exceeds 5000 contacts")
	}
	seen := make(map[string]bool, len(result.Contacts))
	if lookupErr == nil {
		for _, c := range result.Contacts {
			if c.DirectoryID == "" || len(c.DirectoryID) > 8192 || seen[c.DirectoryID] {
				return errors.New("directory returned a missing or duplicate identity")
			}
			seen[c.DirectoryID] = true
			if err := validContact(c); err != nil {
				return err
			}
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw string
	if err = tx.QueryRow(`SELECT value FROM settings WHERE key='directory-profile'`).Scan(&raw); err != nil {
		return err
	}
	var current DirectoryProfile
	if err = json.Unmarshal([]byte(raw), &current); err != nil {
		return err
	}
	if current != profile {
		return errors.New("directory configuration changed during refresh")
	}
	status := DirectoryStatus{}
	err = tx.QueryRow(`SELECT value FROM settings WHERE key='directory-status'`).Scan(&raw)
	if err == nil {
		if err = json.Unmarshal([]byte(raw), &status); err != nil {
			return err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	now := time.Now()
	status.LastAttempt = now
	status.Error = ""
	if lookupErr != nil {
		// Error details can include a bind identity or server diagnostic. Persist a
		// useful category only; credentials and directory replies never enter logs.
		status.Error = "Directory refresh failed. Check the endpoint, credentials and connection; cached contacts remain available."
	} else {
		status.LastSuccess = now
		status.Entries = len(result.Contacts)
		status.Partial = result.Truncated || result.Skipped > 0
		status.Preserved = 0
		for _, c := range result.Contacts {
			var id sql.NullInt64
			var hidden bool
			err = tx.QueryRow(`SELECT contact_id,hidden FROM directory_entries WHERE remote_id=?`, c.DirectoryID).Scan(&id, &hidden)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			payload, err := json.Marshal(c)
			if err != nil {
				return err
			}
			var source string
			if id.Valid {
				if err = tx.QueryRow(`SELECT source FROM contacts WHERE id=?`, id.Int64).Scan(&source); err != nil {
					return err
				}
			}
			if hidden || (id.Valid && source != "directory-sync") {
				status.Preserved++
			} else {
				// A directory may reuse a number or contain duplicate numbers. Never
				// replace a different contact merely because its address now matches.
				var conflict int64
				err = tx.QueryRow(`SELECT id FROM contacts WHERE address=? AND id<>?`, c.Address, id.Int64).Scan(&conflict)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
				if err == nil {
					status.Preserved++
				} else {
					if id.Valid {
						_, err = tx.Exec(`UPDATE contacts SET name=?,address=?,updated=? WHERE id=?`, c.Name, c.Address, now.Format(time.RFC3339Nano), id.Int64)
					} else {
						var inserted sql.Result
						inserted, err = tx.Exec(`INSERT INTO contacts(name,address,source,updated) VALUES(?,?,'directory-sync',?)`, c.Name, c.Address, now.Format(time.RFC3339Nano))
						if err == nil {
							id.Int64, err = inserted.LastInsertId()
							id.Valid = true
						}
					}
					if err != nil {
						return err
					}
					if _, err = tx.Exec(`DELETE FROM contact_numbers WHERE contact_id=?`, id.Int64); err != nil {
						return err
					}
					for i, n := range c.Numbers {
						if _, err = tx.Exec(`INSERT INTO contact_numbers(contact_id,label,address,position) VALUES(?,?,?,?)`, id.Int64, n.Label, n.Address, i); err != nil {
							return err
						}
					}
				}
			}
			if _, err = tx.Exec(`INSERT INTO directory_entries(remote_id,contact_id,payload,hidden) VALUES(?,?,?,?) ON CONFLICT(remote_id) DO UPDATE SET contact_id=excluded.contact_id,payload=excluded.payload`, c.DirectoryID, id, string(payload), hidden); err != nil {
				return err
			}
		}
		if !status.Partial {
			rows, err := tx.Query(`SELECT remote_id,contact_id FROM directory_entries WHERE hidden=0`)
			if err != nil {
				return err
			}
			type removed struct {
				remote string
				id     sql.NullInt64
			}
			missing := []removed{}
			for rows.Next() {
				var item removed
				if err = rows.Scan(&item.remote, &item.id); err != nil {
					rows.Close()
					return err
				}
				if !seen[item.remote] {
					missing = append(missing, item)
				}
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return err
			}
			for _, item := range missing {
				if _, err = tx.Exec(`DELETE FROM directory_entries WHERE remote_id=?`, item.remote); err != nil {
					return err
				}
				if item.id.Valid {
					if _, err = tx.Exec(`DELETE FROM contacts WHERE id=? AND source='directory-sync'`, item.id.Int64); err != nil {
						return err
					}
				}
			}
		}
	}
	encoded, err := json.Marshal(status)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES('directory-status',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, string(encoded)); err != nil {
		return err
	}
	return tx.Commit()
}

// RestoreDirectoryContact discards local name/number changes in favor of the
// latest cached remote entry. Personal notes, favorites and account stay local.
func (s *Store) RestoreDirectoryContact(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw string
	err = tx.QueryRow(`SELECT payload FROM directory_entries WHERE contact_id=? AND hidden=0`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("this contact no longer has a directory entry")
	}
	if err != nil {
		return err
	}
	var contact Contact
	if err = json.Unmarshal([]byte(raw), &contact); err != nil {
		return err
	}
	if err = validContact(contact); err != nil {
		return err
	}
	var conflict int64
	err = tx.QueryRow(`SELECT id FROM contacts WHERE address=? AND id<>?`, contact.Address, id).Scan(&conflict)
	if err == nil {
		return errors.New("another contact already uses the directory number")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err = tx.Exec(`UPDATE contacts SET name=?,address=?,source='directory-sync',updated=? WHERE id=?`, contact.Name, contact.Address, time.Now().Format(time.RFC3339Nano), id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM contact_numbers WHERE contact_id=?`, id); err != nil {
		return err
	}
	for i, n := range contact.Numbers {
		if _, err = tx.Exec(`INSERT INTO contact_numbers(contact_id,label,address,position) VALUES(?,?,?,?)`, id, n.Label, n.Address, i); err != nil {
			return err
		}
	}
	return tx.Commit()
}
