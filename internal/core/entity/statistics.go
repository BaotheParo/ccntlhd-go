package entity

type EventStatistics struct {
	Total     int64 `json:"total"`
	Draft     int64 `json:"draft"`
	Published int64 `json:"published"`
	Cancelled int64 `json:"cancelled"`
	Ended     int64 `json:"ended"`
}

type DashboardStatsResponse struct {
	TotalOrders      int     `json:"total_orders"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalTicketsSold int     `json:"total_tickets_sold"`
}
