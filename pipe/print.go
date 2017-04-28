package pipe

// Println is a log.Println-like function that can be used for debug purposes.
type Println func(...interface{})

// Println logs if not nil
func (p Println) Println(arg ...interface{}) {
	if p != nil {
		p(arg...)
	}
}
