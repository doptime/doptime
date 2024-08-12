package httpserve

func sliceToInterface[T any](slice []T) []interface{} {
	var result []interface{}
	for _, v := range slice {
		result = append(result, v)
	}
	return result
}
