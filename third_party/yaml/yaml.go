package yaml

// Marshal is a stub to satisfy cobra's module dependency in offline builds.
func Marshal(in any) ([]byte, error) {
	return []byte(""), nil
}

// Unmarshal is a stub to satisfy cobra's module dependency in offline builds.
func Unmarshal(data []byte, out any) error {
	return nil
}
