package main

type styleMap map[string]any

func (s styleMap) IsUnset(key string) bool {
	value, ok := s[key]
	if !ok {
		return true
	}
	return isTodoValue(value)
}

func (s styleMap) HasValue(key string) bool {
	value, ok := s[key]
	if !ok {
		return false
	}
	return !isTodoValue(value)
}

func (s styleMap) SetDefault(key, value string) {
	if !s.IsUnset(key) {
		return
	}
	s[key] = value
}

func (s styleMap) SetDefaults(keys []string, value string) {
	for _, key := range keys {
		s.SetDefault(key, value)
	}
}

func (s styleMap) SetRole(key, value string) {
	if value == "" {
		return
	}
	if s.HasValue(key) {
		return
	}
	s[key] = value
}

func (s styleMap) SetAny(key string, value any) {
	if s.HasValue(key) {
		return
	}
	s[key] = value
}

func isTodoValue(value any) bool {
	s, ok := value.(string)
	return ok && s == todoValue
}
