package main

import "log"

func database_test() {
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "postgres"
	dbname := "testdb"

	db, err := createDatabaseConnection(host, port, user, password, dbname)
	if err != nil {
		log.Fatalf("Error creating database connection: %v", err)
	}
	defer db.Close()

	err = startDatabaseMigrations(user, password, host, port)
	if err != nil {
		log.Fatalf("Error running database migrations: %v", err)
	}

	// Create
	newItem := Item{Name: "Sample Item", Description: "This is a sample item"}
	log.Println("Creating new item:", newItem)
	err = createItem(db, newItem)
	if err != nil {
		log.Fatalf("Error creating item: '%v'", err)
	}
	log.Println("Item created successfully:", newItem)

	// Read
	log.Println("Fetching items...")
	items, err := getItems(db)
	if err != nil {
		log.Fatalf("Error fetching items: %v", err)
	}
	log.Println("Items fetched:", items)

	// Update
	updateItem := Item{ID: 1, Name: "Updated Item", Description: "This item has been updated"}
	log.Println("Updating item:", updateItem)
	err = updateItemByID(db, updateItem)
	if err != nil {
		log.Fatalf("Error updating item: %v", err)
	}
	log.Println("Item updated successfully:", updateItem)

	// Delete
	log.Println("Deleting item with ID:", 1)
	err = deleteItemByID(db, 1)
	if err != nil {
		log.Fatalf("Error deleting item: %v", err)
	}
	log.Println("Item deleted successfully")
}
