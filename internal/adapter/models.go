package adapter

// Launch - структура данных для запуска
type Launch struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	ProjectID        int    `json:"projectId"`
	CreatedDate      int64  `json:"createdDate"`
	LastModifiedDate int64  `json:"lastModifiedDate"`
}

// PDFReport - структура данных для PDF-отчета
type PDFReport struct {
	ID          int64  `json:"id"`
	ProjectID   int    `json:"projectId"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	StorageKey  string `json:"storageKey"`
	Name        string `json:"name"`
	CreatedDate int64  `json:"createdDate"`
}
