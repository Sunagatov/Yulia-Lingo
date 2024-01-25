package migrations

import "Yulia-Lingo/internal/db"

func downMigrations() {
	connectDB := db.CreateDatabaseConnection()
	db.DownMigrations(connectDB)
}
