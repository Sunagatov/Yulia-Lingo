package main

import "database/sql"

// createItem inserts a new item into the database
func createItem(db *sql.DB, item Item) error {
	sqlStatement := `INSERT INTO items (name, description) VALUES ($1, $2)`
	_, err := db.Exec(sqlStatement, item.Name, item.Description)
	return err
}

// getItems retrieves all items from the database
func getItems(db *sql.DB) ([]Item, error) {
	rows, err := db.Query("SELECT id, name, description FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var i Item
		if err := rows.Scan(&i.ID, &i.Name, &i.Description); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

// updateItemByID updates an item based on its ID
func updateItemByID(db *sql.DB, item Item) error {
	sqlStatement := `UPDATE items SET name = $2, description = $3 WHERE id = $1`
	_, err := db.Exec(sqlStatement, item.ID, item.Name, item.Description)
	return err
}

// deleteItemByID deletes an item based on its ID
func deleteItemByID(db *sql.DB, id int) error {
	sqlStatement := `DELETE FROM items WHERE id = $1`
	_, err := db.Exec(sqlStatement, id)
	return err
}
