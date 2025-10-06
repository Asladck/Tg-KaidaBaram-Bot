package models

type Statistic struct {
	ID    int64  `db:"id"`
	Event string `db:"event"`
	Data  string `db:"data"`
}
