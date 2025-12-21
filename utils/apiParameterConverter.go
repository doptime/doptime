package utils

import (
	"github.com/doptime/doptime/utils/mapper"
)

func MapToStructDecoder(pIn interface{}) (decoder *mapper.Decoder, err error) {
	// mapstructure support type conversion
	config := &mapper.DecoderConfig{
		Result:           pIn,
		TagName:          "json",
		WeaklyTypedInput: true,
	}

	if decoder, err = mapper.NewDecoder(config); err != nil {
		return nil, err
	}
	return decoder, nil
}
