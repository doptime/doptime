package api

import (
	"encoding/json"
	"reflect"

	"github.com/doptime/doptime/utils"
	"github.com/vmihailenco/msgpack/v5"
)

func (a *ApiCtx[i, o]) GetName() string {
	return a.Name
}
func (a *ApiCtx[i, o]) GetDataSource() string {
	return a.ApiSourceRds
}

func (a *ApiCtx[i, o]) CallByMap(_map map[string]interface{}, msgpackNonstruct []byte, jsonpackNostruct []byte) (ret interface{}, err error) {
	var (
		in          i
		pIn         interface{}
		isTypeInPtr bool = false
		//datapack DataPacked
	)
	// case double pointer decoding
	if vType := reflect.TypeOf((*i)(nil)).Elem(); vType.Kind() == reflect.Ptr {
		pIn = reflect.New(vType.Elem()).Interface()
		in = pIn.(i)
		isTypeInPtr = true
	} else {
		pIn = reflect.New(vType).Interface()
	}

	if len(msgpackNonstruct) > 0 {
		err = msgpack.Unmarshal(msgpackNonstruct, pIn)
	} else if len(jsonpackNostruct) > 0 {
		err = json.Unmarshal(jsonpackNostruct, pIn)
	} else if decoder, errMapTostruct := utils.MapToStructDecoder(pIn); errMapTostruct != nil {
		return nil, errMapTostruct
	} else {
		err = decoder.Decode(_map)
	}

	if err != nil {
		return nil, err
	} else if !isTypeInPtr {
		in = *pIn.(*i)
	}

	//load fill the left fields from db
	if a.ParamEnhancer != nil {
		if out, err := a.ParamEnhancer(in); err != nil {
		} else if isTypeInPtr {
			pIn = out
		} else {
			*pIn.(*i) = out
		}
	}

	//validate the input if it is struct and has tag "validate"
	if err = a.Validate(pIn); err != nil {
		return nil, err
	}
	//post save the result to db
	ret, err = a.Func(in)
	if a.ResultSaver != nil && err == nil {
		_ = a.ResultSaver(in, ret.(o))
	}
	//modify the result value to the web client.
	if a.ResponseModifier != nil {
		ret, err = a.ResponseModifier(in, ret.(o))
	}
	return ret, err
}
