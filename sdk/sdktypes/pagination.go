package sdktypes

type PaginationResult struct {
	TotalCount    int
	NextPageToken string
}

type PaginationRequest struct {
	PageSize  int32
	Skip      int32
	PageToken string
}
