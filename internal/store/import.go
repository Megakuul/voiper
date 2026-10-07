package store

import (
	"fmt"
	"strings"
	"time"
)

// ImportContacts validates everything before writing and leaves matching local
// entries untouched. Shared display names never imply a shared identity.
func (s *Store) ImportContacts(contacts []Contact) (int, error) {
	if len(contacts) > 10000 {
		return 0, fmt.Errorf("import exceeds 10000 contacts")
	}
	for i := range contacts {
		c := &contacts[i]
		c.Address = strings.TrimSpace(c.Address)
		c.Name = strings.TrimSpace(c.Name)
		if c.Name == "" {
			c.Name = c.Address
		}
		if c.Source == "" {
			c.Source = "import"
		}
		if err := validContact(*c); err != nil {
			return 0, fmt.Errorf("contact %d: %w", i+1, err)
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	count := 0
	for _, c := range contacts {
		result, err := tx.Exec(`INSERT INTO contacts(name,address,notes,source,favorite,account,updated) VALUES(?,?,?,?,?,?,?) ON CONFLICT(address) DO NOTHING`, c.Name, c.Address, c.Notes, c.Source, c.Favorite, c.Account, time.Now().Format(time.RFC3339Nano))
		if err != nil {
			return 0, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		if n == 0 {
			continue
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		for i, number := range c.Numbers {
			if _, err = tx.Exec(`INSERT INTO contact_numbers(contact_id,label,address,position) VALUES(?,?,?,?)`, id, number.Label, strings.TrimSpace(number.Address), i); err != nil {
				return 0, err
			}
		}
		count++
	}
	return count, tx.Commit()
}
