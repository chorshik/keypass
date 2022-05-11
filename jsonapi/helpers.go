package jsonapi

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

func regexSafeLower(str string) string {
	return regexp.QuoteMeta(strings.ToLower(str))
}

func isPublicSuffix(host string) bool {
	suffix, _ := publicsuffix.PublicSuffix(host)
	return host == suffix
}

func (s *API) confirmRecipients(name string, recipients []string) ([]string, error) {
	if s.Store.NoConfirm {
		return recipients, nil
	}

	return recipients, fmt.Errorf("user aborted")
}
