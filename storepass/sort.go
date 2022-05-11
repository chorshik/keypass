package storepass

type byLen []string

func (s byLen) Len() int { return len(s) }

func (s byLen) Less(i, j int) bool { return len(s[i]) > len(s[j]) }

func (s byLen) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
