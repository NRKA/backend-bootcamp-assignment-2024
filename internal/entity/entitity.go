package entity

type Status string

const (
	Created      Status = "created"
	Approved     Status = "approved"
	Declined     Status = "declined"
	OnModeration Status = "on moderation"
)

type Flat struct {
	ID      int    `json:"id"`
	HouseID int    `json:"house_id"`
	Price   int    `json:"price"`
	Rooms   int    `json:"rooms"`
	Status  Status `json:"status"`
}
