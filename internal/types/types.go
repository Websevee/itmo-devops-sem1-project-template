package types

type ArchiveType string

const (
	Zip ArchiveType = "zip"
	Tar ArchiveType = "tar"
)

type Product struct {
	Id        int
	ProductId int
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
