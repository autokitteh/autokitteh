package dbgorm

func (db *gormdb) Report() []error {
	s, _ := db.db.DB()
	err := s.Ping()
	return []error{err}
}
