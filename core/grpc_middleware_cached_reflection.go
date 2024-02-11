package core

import (
	"reflect"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

var cachedReflectionMiddlewares = make(map[string]map[string][]string)
var _cachedReflectionMiddlewares sync.RWMutex

// concurrenty context
func getCachedMiddlewareByServiceEndpoint(info *grpc.UnaryServerInfo) []string {
	methodParts := strings.Split(info.FullMethod, "/")
	service := methodParts[1]
	methodName := methodParts[len(methodParts)-1]

	_cachedReflectionMiddlewares.RLock()
	result, exists := cachedReflectionMiddlewares[service]
	_cachedReflectionMiddlewares.RUnlock()
	if !exists {
		v := reflect.ValueOf(info.Server)
		if middlewaresReflection, found := extractMiddlewares(v); found {
			middlewaresMap := convertReflectValueToMap(middlewaresReflection)
			_cachedReflectionMiddlewares.Lock()
			cachedReflectionMiddlewares[service] = middlewaresMap
			_cachedReflectionMiddlewares.Unlock()

			return middlewaresMap[methodName]
		}
	}
	return result[methodName]
}

func convertReflectValueToMap(value reflect.Value) map[string][]string {
	if value.Kind() == reflect.Interface {
		value = value.Elem()
	}

	if value.Kind() != reflect.Map {
		return nil
	}

	result := make(map[string][]string)

	for _, key := range value.MapKeys() {
		if key.Kind() != reflect.String {
			continue
		}

		val := value.MapIndex(key)
		if val.Kind() == reflect.Interface {
			val = val.Elem()
		}

		if val.Kind() == reflect.Slice {
			var strs []string
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i)
				if elem.Kind() == reflect.Interface {
					elem = elem.Elem()
				}
				if elem.Kind() == reflect.String {
					strs = append(strs, elem.String())
				} else {
					return nil
				}
			}
			result[key.String()] = strs
		} else {
			return nil
		}
	}

	return result
}

func extractMiddlewares(v reflect.Value) (reflect.Value, bool) {
	var findMiddlewares func(reflect.Value) (reflect.Value, bool)
	findMiddlewares = func(value reflect.Value) (reflect.Value, bool) {
		if value.Kind() == reflect.Ptr && !value.IsNil() {
			value = value.Elem()
		}
		if value.Kind() == reflect.Struct {
			for i := 0; i < value.NumField(); i++ {
				field := value.Field(i)

				if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
					if result, found := findMiddlewares(field); found {
						return result, true
					}
				} else if field.Kind() == reflect.Map {
					for _, key := range field.MapKeys() {
						if key.String() == "middlewares" {
							return field.MapIndex(key), true
						}
					}
				}
			}
		}
		return reflect.Value{}, false
	}

	return findMiddlewares(v)
}
