package main
import (
    "context"
    "fmt"
    "Yulia-Lingo/internal/config"
    "Yulia-Lingo/internal/database"
    "Yulia-Lingo/internal/logger"
)
func main() {
    cfg, _ := config.Load()
    cfg.Telegram.BotToken = "dummy"
    lg := logger.New(cfg)
    db, _ := database.Connect(context.Background(), cfg, lg)
    defer db.Close()
    
    rows, _ := db.Query(context.Background(), `
        SELECT word, preposition, COUNT(*) as cnt 
        FROM words 
        WHERE user_id = 591545118 
        GROUP BY word, preposition 
        HAVING COUNT(*) > 1 
        ORDER BY cnt DESC, word 
        LIMIT 10
    `)
    fmt.Println("🔍 Sample duplicates (word + preposition):")
    for rows.Next() {
        var word, prep string
        var count int
        rows.Scan(&word, &prep, &count)
        if prep == "" {
            fmt.Printf("   '%s' (no prep): %d times\n", word, count)
        } else {
            fmt.Printf("   '%s %s': %d times\n", word, prep, count)
        }
    }
    
    rows2, _ := db.Query(context.Background(), `
        SELECT word, COUNT(DISTINCT preposition) as prep_count
        FROM words 
        WHERE user_id = 591545118 
        GROUP BY word 
        HAVING COUNT(DISTINCT preposition) > 1 
        ORDER BY COUNT(DISTINCT preposition) DESC 
        LIMIT 10
    `)
    fmt.Println("\n📝 Words with multiple prepositions:")
    for rows2.Next() {
        var word string
        var count int
        rows2.Scan(&word, &count)
        fmt.Printf("   '%s': %d variations\n", word, count)
    }
}
