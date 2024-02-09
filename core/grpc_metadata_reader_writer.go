package core

import "google.golang.org/grpc/metadata"

// metadataReaderWriter sert de pont entre les métadonnées gRPC et le propagateur OpenTelemetry.
type MetadataReaderWriter struct {
	metadata.MD
}

func (mrw MetadataReaderWriter) Get(key string) string {
	values := mrw.MD[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (mrw MetadataReaderWriter) Set(key string, value string) {
	mrw.MD[key] = []string{value}
}

func (mrw MetadataReaderWriter) Keys() []string {
	keys := make([]string, 0, len(mrw.MD))
	for k := range mrw.MD {
		keys = append(keys, k)
	}
	return keys
}

func (mrw MetadataReaderWriter) ForeachKey(handler func(key string, val string) error) error {
	for k, vals := range mrw.MD {
		for _, val := range vals {
			if err := handler(k, val); err != nil {
				return err
			}
		}
	}
	return nil
}
