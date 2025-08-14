package domain

type Reservation struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"` // e.g., "confirmed", "cancelled"
}
