package action

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Complete ...
func (s *Action) Complete(*cli.Context) {
	list, err := s.Store.List()
	if err != nil {
		return
	}

	for _, v := range list {
		fmt.Println(v)
	}
}
