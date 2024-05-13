package dbgorm

func (db *gormdb) Report() error {
	s, _ := db.db.DB()
	return s.Ping()
}
