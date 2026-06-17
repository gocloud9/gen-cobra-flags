package multi

// AlphaRequest is the first annotated struct.
// +cobra:flag=alpha
// +cobra:short=a
// +cobra:usage=Alpha request
type AlphaRequest struct {
	// +cobra:flag=alpha-name
	// +cobra:usage=Alpha name
	// +cobra:default=""
	Name string
}

// BetaRequest is the second annotated struct.
// +cobra:flag=beta
// +cobra:short=b
// +cobra:usage=Beta request
type BetaRequest struct {
	// +cobra:flag=beta-size
	// +cobra:usage=Beta size
	// +cobra:default=0
	Size int
}
