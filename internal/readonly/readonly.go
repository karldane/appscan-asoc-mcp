package readonly

type ReadOnlyChecker interface {
	ReadOnly() bool
}

func IsReadOnly(cfg ReadOnlyChecker) bool {
	if cfg == nil {
		return false
	}
	return cfg.ReadOnly()
}
