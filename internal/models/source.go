package models

import "time"

type Source struct {
	ID         int64
	Name       string
	FeedURL    string
	TypeSource int
	Priority   int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
