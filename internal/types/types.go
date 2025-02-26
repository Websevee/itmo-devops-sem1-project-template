package types

type ArchiveType int

const (
	Zip ArchiveType = iota // 0
	Tar                    // 1
)

type Product struct {
	Id        int
	CreatedAt string
	Name      string
	Category  string
	Price     float64
}

type GetPricesResponse struct {
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}
