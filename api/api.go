package api

import (
	"context"
	"reflect"
	"strconv"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/msgpack/v5"
)

// crate ApiFun. the created New can be used as normal function:
//
//	f := func(InParam *InDemo) (ret string, err error) , this is logic function
//	options. there are 2 poosible options:
//		1. api.Name("ServiceName")  //set the ServiceName of the New. which is string. default is the name of the InParameter type but with "In" removed
//		2. api.DB("RedisDatabaseName")  //set the DB name of the job. default is the name of the function
//
// ServiceName is defined as "In" + ServiceName in the InParameter
// ServiceName is automatically converted to lower case
func New[i any, o any](f func(InParameter i) (ret o, err error), options ...ApiOption) (retf func(InParam i) (ret o, err error)) {
	var (
		option   *ApiOption = &ApiOption{DataSource: "default"}
		validate            = needValidate(reflect.TypeOf(new(i)).Elem())
		decoder  *mapstructure.Decoder
	)
	if len(options) > 0 {
		option = &options[0]
	}

	if len(option.Name) > 0 {
		option.Name = specification.ApiName(option.Name)
	}
	if len(option.Name) == 0 {
		option.Name = specification.ApiNameByType((*i)(nil))
	}
	if len(option.Name) == 0 {
		log.Error().Str("service misnamed", option.Name).Send()
	}

	if _, ok := specification.DisAllowedServiceNames[option.Name]; ok {
		log.Error().Str("service misnamed", option.Name).Send()
	}
	//warn if DataSource not defined in the environment
	//however, the DataSource may be specified later in the environment, by dynamic loading from remove config file
	if _, err := config.GetRdsClientByName(option.DataSource); err != nil {
		log.Warn().Str("this data source specified in the api is not defined in the environment. Please check the configuration", option.DataSource).Send()
	}

	log.Debug().Str("Api service create start. name", option.Name).Send()
	//create a goroutine to process one job
	ProcessOneJob := func(s []byte) (ret interface{}, err error) {
		var (
			in   i
			pIn  interface{}
			_map map[string]interface{} = map[string]interface{}{}
			//datapack DataPacked
		)
		// case double pointer decoding
		if vType := reflect.TypeOf((*i)(nil)).Elem(); vType.Kind() == reflect.Ptr {
			pIn = reflect.New(vType.Elem()).Interface()
			in = pIn.(i)
		} else {
			pIn = reflect.New(vType).Interface()
			in = *pIn.(*i)
		}

		//type conversion of form data (from url parameter or post form)
		if err = msgpack.Unmarshal(s, &_map); err != nil {
			return nil, err
		}
		hook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
			switch {
			case f.Kind() == reflect.String && t.Kind() == reflect.Int64:
				return strconv.ParseInt(data.(string), 10, 64)
			case f.Kind() == reflect.String && t.Kind() == reflect.Int:
				return strconv.Atoi(data.(string))
			case f.Kind() == reflect.String && t.Kind() == reflect.Int32:
				i, err := strconv.ParseInt(data.(string), 10, 32)
				return int32(i), err

			case f.Kind() == reflect.String && t.Kind() == reflect.Float64:
				return strconv.ParseFloat(data.(string), 64)
			case f.Kind() == reflect.String && t.Kind() == reflect.Float32:
				f, err := strconv.ParseFloat(data.(string), 32)
				if err != nil {
					return nil, err
				}
				return float32(f), nil

			case f.Kind() == reflect.Int64 && t.Kind() == reflect.String:
				return strconv.FormatInt(data.(int64), 10), nil
			case f.Kind() == reflect.Int && t.Kind() == reflect.String:
				return strconv.Itoa(data.(int)), nil
			case f.Kind() == reflect.Int32 && t.Kind() == reflect.String:
				return strconv.FormatInt(int64(data.(int32)), 10), nil

			case f.Kind() == reflect.Float64 && t.Kind() == reflect.String:
				return strconv.FormatFloat(data.(float64), 'f', -1, 64), nil
			case f.Kind() == reflect.Float32 && t.Kind() == reflect.String:
				return strconv.FormatFloat(float64(data.(float32)), 'f', -1, 32), nil
			case f.Kind() == reflect.Bool && t.Kind() == reflect.String:
				return strconv.FormatBool(data.(bool)), nil
			default:
				return data, nil
			}
		}

		//mapstructure support type conversion
		config := &mapstructure.DecoderConfig{
			Metadata:   nil,
			Result:     pIn,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(hook),
		}

		if decoder, err = mapstructure.NewDecoder(config); err != nil {
			return nil, err
		}
		if err = decoder.Decode(_map); err != nil {
			return nil, err
		}
		//validate the input if it is struct and has tag "validate"
		if err = validate(pIn); err != nil {
			return nil, err
		}
		return f(in)
	}
	//register Api
	apiInfo := &ApiInfo{
		Name:                      option.Name,
		DataSource:                option.DataSource,
		WithHeader:                HeaderFieldsUsed(reflect.TypeOf(new(i)).Elem()),
		WithJwt:                   WithJwtFields(reflect.TypeOf(new(i)).Elem()),
		ApiFuncWithMsgpackedParam: ProcessOneJob,
		Ctx:                       context.Background(),
	}
	ApiServices.Set(option.Name, apiInfo)
	funcPtr := reflect.ValueOf(f).Pointer()
	fun2ApiInfoMap.Store(funcPtr, apiInfo)
	APIGroupByDataSource.Upsert(option.DataSource, []string{}, func(exist bool, valueInMap, newValue []string) []string {
		return append(valueInMap, option.Name)
	})
	log.Debug().Str("ApiNamed service created completed!", option.Name).Send()
	//return Api context
	return f
}
