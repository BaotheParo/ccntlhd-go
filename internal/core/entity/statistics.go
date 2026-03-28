package entity

type EventStatistics struct {
	Total     int64 `json:"total"`
	Draft     int64 `json:"draft"`
	Published int64 `json:"published"`
	Cancelled int64 `json:"cancelled"`
	Ended     int64 `json:"ended"`
}
