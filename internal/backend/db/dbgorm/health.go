package dbgorm

import "fmt"

func (db *gormdb) Report() error {
	w, err := db.wdb.DB()
	if err != nil {
		return fmt.Errorf("write db: %w", err)
	}

	if err := w.Ping(); err != nil {
		return fmt.Errorf("write db ping: %w", err)
	}

	r, err := db.wdb.DB()
	if err != nil {
		return fmt.Errorf("read db: %w", err)
	}

	if err := r.Ping(); err != nil {
		return fmt.Errorf("read db ping: %w", err)
	}

	return nil
}
