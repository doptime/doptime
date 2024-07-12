package httpserve

func convertKeysToBytes(keys []string) (keyvals [][]byte) {
	keyvals = make([][]byte, 0, len(keys)+1)
	for i := 0; i < len(keys); i += 1 {
		keyvals = append(keyvals, []byte(keys[i]))
	}
	return keyvals
}
