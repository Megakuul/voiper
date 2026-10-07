package store

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Store) Contacts(query string) ([]Contact, error) { return s.ContactsPage(query, 0, 500) }
func (s *Store) ContactsPage(query string, offset, limit int) ([]Contact, error) {
	if offset < 0 {
		return nil, errors.New("invalid contact offset")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT id,name,address,notes,source,favorite,account,updated,
COALESCE((SELECT remote_id FROM directory_entries WHERE contact_id=contacts.id LIMIT 1),'') FROM contacts
WHERE instr(lower(name),lower(?))>0 OR instr(lower(address),lower(?))>0 OR EXISTS (SELECT 1 FROM contact_numbers n WHERE n.contact_id=contacts.id AND instr(lower(n.address),lower(?))>0)
ORDER BY favorite DESC,name COLLATE NOCASE,id LIMIT ? OFFSET ?`, query, query, query, limit, offset)
	if err != nil {
		return nil, err
	}
	contacts := []Contact{}
	for rows.Next() {
		var c Contact
		var updated string
		if err = rows.Scan(&c.ID, &c.Name, &c.Address, &c.Notes, &c.Source, &c.Favorite, &c.Account, &updated, &c.DirectoryID); err != nil {
			rows.Close()
			return nil, err
		}
		c.Updated, _ = time.Parse(time.RFC3339Nano, updated)
		c.Numbers = []ContactNumber{}
		contacts = append(contacts, c)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if len(contacts) == 0 {
		return contacts, nil
	}
	ids := make([]any, len(contacts))
	positions := map[int64]int{}
	for i, c := range contacts {
		ids[i] = c.ID
		positions[c.ID] = i
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	rows, err = s.db.Query(`SELECT contact_id,label,address FROM contact_numbers WHERE contact_id IN (`+placeholders+`) ORDER BY position`, ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var n ContactNumber
		if err = rows.Scan(&id, &n.Label, &n.Address); err != nil {
			return nil, err
		}
		i := positions[id]
		contacts[i].Numbers = append(contacts[i].Numbers, n)
	}
	return contacts, rows.Err()
}
func validContact(c Contact) error {
	if strings.TrimSpace(c.Address) == "" || len(c.Address) > 1024 || strings.ContainsAny(c.Address, "\r\n\x00") {
		return errors.New("enter a valid number or SIP address")
	}
	if len(c.Name) > 512 || len(c.Notes) > 8192 || len(c.Account) > 256 {
		return errors.New("contact text is too long")
	}
	if len(c.Numbers) > 20 {
		return errors.New("a contact supports at most 20 additional numbers")
	}
	seen := map[string]bool{strings.TrimSpace(c.Address): true}
	for _, n := range c.Numbers {
		address := strings.TrimSpace(n.Address)
		if address == "" || len(address) > 1024 || strings.ContainsAny(address, "\r\n\x00") || len(n.Label) > 80 {
			return errors.New("enter a valid additional number and label")
		}
		if seen[address] {
			return fmt.Errorf("duplicate contact number %q", address)
		}
		seen[address] = true
	}
	return nil
}
func (s *Store) SaveContact(c Contact) error {
	c.Address = strings.TrimSpace(c.Address)
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		c.Name = c.Address
	}
	if err := validContact(c); err != nil {
		return err
	}
	if c.Source == "" {
		c.Source = "local"
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	updated := time.Now().Format(time.RFC3339Nano)
	if c.ID != 0 {
		result, err := tx.Exec(`UPDATE contacts SET name=?,address=?,notes=?,favorite=?,account=?,updated=?,source=CASE WHEN source='directory-sync' THEN 'local' ELSE source END WHERE id=?`, c.Name, c.Address, c.Notes, c.Favorite, c.Account, updated, c.ID)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed == 0 {
			return errors.New("contact no longer exists")
		}
	} else {
		_, err = tx.Exec(`INSERT INTO contacts(name,address,notes,source,favorite,account,updated) VALUES(?,?,?,?,?,?,?) ON CONFLICT(address) DO UPDATE SET name=excluded.name,notes=excluded.notes,favorite=excluded.favorite,account=excluded.account,updated=excluded.updated,source=CASE WHEN contacts.source='directory-sync' THEN 'local' ELSE contacts.source END`, c.Name, c.Address, c.Notes, c.Source, c.Favorite, c.Account, updated)
		if err != nil {
			return err
		}
		if err = tx.QueryRow(`SELECT id FROM contacts WHERE address=?`, c.Address).Scan(&c.ID); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(`DELETE FROM contact_numbers WHERE contact_id=?`, c.ID); err != nil {
		return err
	}
	for i, n := range c.Numbers {
		if _, err = tx.Exec(`INSERT INTO contact_numbers(contact_id,label,address,position) VALUES(?,?,?,?)`, c.ID, strings.TrimSpace(n.Label), strings.TrimSpace(n.Address), i); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (s *Store) DeleteContact(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE directory_entries SET hidden=1 WHERE contact_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM contacts WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
