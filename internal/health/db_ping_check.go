package health

import "context"

// DBPinger is a minimal interface for databases that support a Ping operation
// (e.g. *sql.DB implements PingContext; wrap it to match this interface).
type DBPinger interface {
	Ping(ctx context.Context) error
}

type DBPingCheck struct {
	name     string
	critical bool
	db       DBPinger
}

func NewDBPingCheck(name string, critical bool, db DBPinger) ReadinessCheck {
	return &DBPingCheck{name: name, critical: critical, db: db}
}

func (c *DBPingCheck) Name() string   { return c.name }
func (c *DBPingCheck) Critical() bool { return c.critical }

func (c *DBPingCheck) Check(ctx context.Context) error {
	return c.db.Ping(ctx)
}
