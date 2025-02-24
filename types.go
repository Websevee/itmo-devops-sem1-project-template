package main

type Product struct {
	ID       int
	Created  string
	Name     string
	Category string
	Price    float64
}

type Response struct {
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}
