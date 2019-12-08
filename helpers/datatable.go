package helpers

// DatatableResponse object
type DatatableResponse struct {
	Draw            int        `json:"draw"`
	RecordsTotal    int        `json:"recordsTotal"`
	RecordsFiltered int        `json:"recordsFiltered"`
	Data            [][]string `json:"data"`
}

// DatatableRequest object
type DatatableRequest struct {
	Parameters Parameters `json:"parameters"`
}

// Parameters object
type Parameters struct {
	Draw   int     `json:"draw"`
	Length int     `json:"length"`
	Start  int     `json:"start"`
	Search Search  `json:"search"`
	Order  []Order `json:"order"`
}

// Search object
type Search struct {
	Regex bool   `json:"regex"`
	Value string `json:"value"`
}

// Order object
type Order struct {
	Column    int    `json:"column"`
	Direction string `json:"dir"`
}
