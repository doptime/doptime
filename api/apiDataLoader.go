package api

import (
	"strings"
)

func DataMerger(des map[string]interface{}, src map[string]interface{}) {
	//copy fields from out to _mp
	for key, value := range src {
		//skip the key that already exists in des
		if _, ok := des[key]; ok {
			continue
		}
		des[key] = value
	}
}
func KeyFieldReplacedWithJwt(k string, f string, _mp map[string]interface{}) (string, string) {
	var key, field string = k, f
	for key, value := range _mp {
		if vstr, ok := value.(string); ok {
			if strings.Contains(key, "@"+key) {
				key = strings.Replace(key, "@"+key, vstr, 1)
			}
			if strings.Contains(field, "@"+key) {
				field = strings.Replace(field, "@"+key, vstr, 1)
			}
		}
	}
	return key, field
}
