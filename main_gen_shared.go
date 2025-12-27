// Generated Code - Do Not Edit
package todo




var defaultAdaptors map[string]func(any) any

func RegisterAdaptor(adaptors ...func(any) any) {
  for _, adaptor := range adaptors {
    defaultAdaptors = append(defaultAdaptors, adaptor)
  }
}