package my_word_list

type Category struct {
	Name  string
	Emoji string
}

var DefaultCategories = []Category{
	{Name: "Travel & Places", Emoji: "✈️"},
	{Name: "Food & Drinks", Emoji: "🍕"},
	{Name: "Work & Business", Emoji: "💼"},
	{Name: "Emotions & Feelings", Emoji: "❤️"},
	{Name: "Home & Daily Life", Emoji: "🏠"},
	{Name: "Hobbies & Interests", Emoji: "🎨"},
	{Name: "Health & Body", Emoji: "🏥"},
	{Name: "People & Relationships", Emoji: "👥"},
	{Name: "Nature & Environment", Emoji: "🌍"},
	{Name: "Education & Learning", Emoji: "📚"},
	{Name: "Money & Shopping", Emoji: "💰"},
	{Name: "Technology", Emoji: "⚙️"},
	{Name: "Entertainment", Emoji: "🎭"},
	{Name: "Transportation", Emoji: "🚗"},
	{Name: "Communication", Emoji: "📱"},
	{Name: "Other", Emoji: "⚡"},
}

func GetCategoryEmoji(categoryName string) string {
	for _, cat := range DefaultCategories {
		if cat.Name == categoryName {
			return cat.Emoji
		}
	}
	return "⚡"
}
