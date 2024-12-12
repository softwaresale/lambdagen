package pkg

// RequestVarConvert is an interface for converting string request variables into custom user types.
// These custom user types should still be based on literals.
type RequestVarConvert[ResT any] func(rawVal string) (ResT, error)
