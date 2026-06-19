package flagadaptor

// FilterRequest exercises a custom flag adaptor that converts a repeatable
// StringArray flag into a map config field.
// +cobra:flag=filter-config
// +cobra:short=c
// +cobra:usage=Filter configuration
type FilterRequest struct {
	// +cobra:flag=tag
	// +cobra:short=t
	// +cobra:usage=Tag filters. Format is key=value or key.
	// +cobra:flag:type=StringArray
	// +cobra:flag:adaptor=BuildTags
	Tags map[string]string

	// +cobra:flag=region
	// +cobra:usage=Region filter
	Region string
}
