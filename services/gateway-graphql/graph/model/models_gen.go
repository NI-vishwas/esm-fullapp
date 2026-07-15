package model

type Event struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	BannerURL      string `json:"bannerUrl"`
	AvailableSlots int    `json:"availableSlots"`
}

type Booking struct {
	ID           string `json:"id"`
	EventID      string `json:"eventId"`
	UserID       string `json:"userId"`
	SeatsCount   int    `json:"seatsCount"`
	Status       string `json:"status"`
	TicketPdfURL string `json:"ticketPdfUrl"`
}