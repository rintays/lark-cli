package larkapi

type Minute struct {
	Token      string `json:"token"`
	OwnerID    string `json:"owner_id,omitempty"`
	CreateTime string `json:"create_time,omitempty"`
	Title      string `json:"title,omitempty"`
	Cover      string `json:"cover,omitempty"`
	Duration   string `json:"duration,omitempty"`
	URL        string `json:"url,omitempty"`
}

type ListMinutesRequest struct {
	PageSize   int
	PageToken  string
	UserIDType string
}

type ListMinutesResult struct {
	Items     []Minute
	PageToken string
	HasMore   bool
}
