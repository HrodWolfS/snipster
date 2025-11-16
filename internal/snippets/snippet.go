package snippets

import "time"

type Snippet struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Category  string    `json:"category"`
    Language  string    `json:"language"`
    Tags      []string  `json:"tags"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`

    // Path on disk (not serialized)
    Path string `json:"-"`
}

