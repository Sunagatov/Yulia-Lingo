import csv

MULTI_PREPS = sorted([
    "away from", "back against", "back on", "back at", "back from", "back into",
    "back to", "back up", "down into", "down on", "in favor of", "in with",
    "off with", "out of", "out to", "up on", "up to", "up with",
    "clear of", "the question", "the roost", "the roost in",
    "the top spot", "the tone for", "the limit", "the heart",
    "the right direction", "a bond", "a mark on", "a shift in",
    "a target", "a glance at", "an exclusive club", "an insight",
    "an eye on", "up where it left off", "friction in",
    "back at the heart", "one's own against", "in the right direction",
], key=len, reverse=True)

PARTICLES = {
    "for", "from", "to", "on", "in", "into", "with", "at", "up", "out",
    "off", "over", "through", "down", "away", "back", "by", "of", "about",
    "after", "against", "around", "beneath", "toward", "still",
}

def split_verb(word):
    w = word.strip()
    for m in MULTI_PREPS:
        if w.endswith(" " + m):
            base = w[:-(len(m) + 1)].strip()
            return base, m
    parts = w.split()
    if len(parts) == 1:
        return w, ""
    if parts[-1].lower() in PARTICLES:
        return " ".join(parts[:-1]), parts[-1]
    return w, ""

rows = []
with open("resource/import/vocabulary_verb.csv", newline="", encoding="utf-8") as f:
    reader = csv.reader(f)
    for i, row in enumerate(reader):
        if i == 0:
            rows.append(["word", "part_of_speech", "preposition", "translations"])
            continue
        if len(row) < 3:
            continue
        base, prep = split_verb(row[0])
        rows.append([base, row[1], prep, row[2]])

with open("resource/import/vocabulary_verb.csv", "w", newline="", encoding="utf-8") as f:
    writer = csv.writer(f)
    writer.writerows(rows)

print(f"Done: {len(rows)-1} verb rows")
