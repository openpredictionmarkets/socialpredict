package mcpserver

type PageOutput struct {
	Limit      int  `json:"limit" jsonschema:"page size used for this result"`
	Offset     int  `json:"offset" jsonschema:"zero-based offset used for this result"`
	Count      int  `json:"count" jsonschema:"number of returned items"`
	NextOffset *int `json:"nextOffset,omitempty" jsonschema:"offset for the next page when one is known or inferred"`
	HasMore    bool `json:"hasMore" jsonschema:"whether another page is available or likely"`
	Total      *int `json:"total,omitempty" jsonschema:"full total when the underlying service knows it"`
}

type PageItems[T any] struct {
	Items []T        `json:"items" jsonschema:"items on this page"`
	Page  PageOutput `json:"page" jsonschema:"pagination metadata"`
}

func NewPageOutput(limit int, offset int, count int, total *int) PageOutput {
	hasMore := false
	if total != nil {
		hasMore = offset+count < *total
	} else {
		hasMore = count == limit && count > 0
	}
	var next *int
	if hasMore {
		value := offset + count
		next = &value
	}
	return PageOutput{
		Limit:      limit,
		Offset:     offset,
		Count:      count,
		NextOffset: next,
		HasMore:    hasMore,
		Total:      total,
	}
}

func NewPageItems[T any](items []T, limit int, offset int, total *int) PageItems[T] {
	if items == nil {
		items = []T{}
	}
	return PageItems[T]{
		Items: items,
		Page:  NewPageOutput(limit, offset, len(items), total),
	}
}
