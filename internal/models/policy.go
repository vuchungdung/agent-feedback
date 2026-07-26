package models

type Policy struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Rules       []string `json:"rules"`
	Reference   string   `json:"reference"`
}
