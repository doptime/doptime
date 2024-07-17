package api

import (
	cmap "github.com/orcaman/concurrent-map/v2"
)

var ApiProviderMap cmap.ConcurrentMap[string, *PublishOptions] = cmap.New[*PublishOptions]()
