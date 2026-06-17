package simple

// SimpleRequest is a minimal annotated struct.
// +cobra:flag=simple
// +cobra:short=s
// +cobra:usage=A simple request
type SimpleRequest struct {
	// +cobra:flag=title
	// +cobra:short=t
	// +cobra:usage=The title
	// +cobra:default=""
	Title string

	// +cobra:flag=count
	// +cobra:usage=The count
	// +cobra:default=0
	Count int

	// +cobra:flag=enabled
	// +cobra:usage=Whether enabled
	// +cobra:default=false
	Enabled bool
}
