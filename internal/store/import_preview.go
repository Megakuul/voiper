package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ContactImportPreview struct {
	Entries                          []ContactImportEntry
	Added, Updated, Skipped, Invalid int
	Committed                        bool
}

func (s *Store) PreviewContactImport(format, text string, mergeNumbers bool) (ContactImportPreview, error) {
	return s.importContactFile(format, text, mergeNumbers, false)
}

func (s *Store) ImportContactFile(format, text string, mergeNumbers bool) (ContactImportPreview, error) {
	return s.importContactFile(format, text, mergeNumbers, true)
}

func (s *Store) importContactFile(format, text string, mergeNumbers, commit bool) (ContactImportPreview, error) {
	entries, err := parseContactFile(format, text)
	if err != nil {
		return ContactImportPreview{}, err
	}
	preview := ContactImportPreview{Entries: entries}
	tx, err := s.db.Begin()
	if err != nil {
		return preview, err
	}
	defer tx.Rollback()
	seen := map[string]bool{}
	updates := map[int][]ContactNumber{}
	positions := map[int]int{}
	for i := range preview.Entries {
		entry := &preview.Entries[i]
		if entry.Action == "invalid" {
			preview.Invalid++
			continue
		}
		address := entry.Contact.Address
		if seen[address] {
			entry.Action, entry.Detail = "skip", "duplicate primary address in this file; first record takes precedence"
			preview.Skipped++
			continue
		}
		seen[address] = true
		err = tx.QueryRow(`SELECT id FROM contacts WHERE address=?`, address).Scan(&entry.Contact.ID)
		if errors.Is(err, sql.ErrNoRows) {
			entry.Action, entry.Detail = "add", "new primary address"
			preview.Added++
			continue
		}
		if err != nil {
			return preview, err
		}
		entry.Action, entry.Detail = "skip", "existing contact preserved"
		if mergeNumbers && len(entry.Contact.Numbers) > 0 {
			numbers := map[string]bool{address: true}
			rows, err := tx.Query(`SELECT address,position FROM contact_numbers WHERE contact_id=?`, entry.Contact.ID)
			if err != nil {
				return preview, err
			}
			for rows.Next() {
				var number string
				var position int
				if err = rows.Scan(&number, &position); err != nil {
					rows.Close()
					return preview, err
				}
				numbers[number] = true
				positions[i] = max(positions[i], position+1)
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return preview, err
			}
			for _, number := range entry.Contact.Numbers {
				number.Address = strings.TrimSpace(number.Address)
				if !numbers[number.Address] {
					updates[i] = append(updates[i], number)
					numbers[number.Address] = true
				}
			}
			if len(numbers) > 21 {
				entry.Action, entry.Detail = "invalid", "merging would exceed 20 additional numbers; keep existing contacts or edit this record"
				preview.Invalid++
				continue
			}
			if len(updates[i]) > 0 {
				entry.Action, entry.Detail = "update", fmt.Sprintf("add %d numbers; existing name, notes, labels and preferences preserved", len(updates[i]))
				preview.Updated++
				continue
			}
		}
		preview.Skipped++
	}
	if !commit || preview.Invalid > 0 {
		return preview, nil
	}
	updated := time.Now().Format(time.RFC3339Nano)
	for i, entry := range preview.Entries {
		c := entry.Contact
		numbers := updates[i]
		switch entry.Action {
		case "add":
			result, err := tx.Exec(`INSERT INTO contacts(name,address,notes,source,updated) VALUES(?,?,?,?,?)`, c.Name, c.Address, c.Notes, "import", updated)
			if err != nil {
				return preview, err
			}
			c.ID, err = result.LastInsertId()
			if err != nil {
				return preview, err
			}
			numbers = c.Numbers
		case "update":
			if _, err = tx.Exec(`UPDATE contacts SET updated=? WHERE id=?`, updated, c.ID); err != nil {
				return preview, err
			}
		default:
			continue
		}
		for j, number := range numbers {
			if _, err = tx.Exec(`INSERT INTO contact_numbers(contact_id,label,address,position) VALUES(?,?,?,?)`, c.ID, strings.TrimSpace(number.Label), strings.TrimSpace(number.Address), positions[i]+j); err != nil {
				return preview, err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return preview, err
	}
	preview.Committed = true
	return preview, nil
}
