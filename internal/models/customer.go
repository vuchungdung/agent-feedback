package models

type Customer struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Tier        string   `json:"tier"`
	Tenure      string   `json:"tenure"`
	PastTickets int      `json:"past_tickets"`
	TotalOrders int      `json:"total_orders"`
	Tags        []string `json:"tags"`
}
