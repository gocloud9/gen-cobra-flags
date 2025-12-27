// Generated Code - Do Not Edit
package todo



var adaptorMyCustomTypeAdaptor func(string) MyCustomType

func RegisterAdaptorMyCustomTypeAdaptor(f func(string) MyCustomType) {
    adaptorMyCustomTypeAdaptor = f
}

var defaultAdaptors map[string]func(any) any

func RegisterAdaptor(adaptors ...func(any) any) {
  for _, adaptor := range adaptors {
    defaultAdaptors = append(defaultAdaptors, adaptor)
  }
}