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
    
    var total int
    db.QueryRow(context.Background(), "SELECT COUNT(*) FROM words WHERE user_id = 591545118").Scan(&total)
    fmt.Printf("✅ Total words: %d\n\n", total)
    
    var dups int
    db.QueryRow(context.Background(), "SELECT COUNT(*) FROM (SELECT word, COUNT(*) as cnt FROM words WHERE user_id = 591545118 GROUP BY word HAVING COUNT(*) > 1) sub").Scan(&dups)
    fmt.Printf("✅ Duplicate words: %d\n\n", dups)
    
    rows, _ := db.Query(context.Background(), "SELECT part_of_speech, COUNT(*) FROM word_meanings wm JOIN words w ON w.id = wm.word_id WHERE w.user_id = 591545118 GROUP BY part_of_speech ORDER BY COUNT(*) DESC")
    fmt.Println("📊 Words by part of speech:")
    for rows.Next() {
        var pos string
        var count int
        rows.Scan(&pos, &count)
        if pos != "" {
            fmt.Printf("   %s: %d\n", pos, count)
        }
    }
}
