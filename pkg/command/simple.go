package command

import flag "github.com/spf13/pflag"

type simple struct {
	synopsisVal string
	helpVal     string
	runFunc     func([]string) error
}

func (s *simple) run(args []string) error {
	return s.runFunc(args)
}
func (s *simple) synopsis() string {
	return s.synopsisVal
}
func (s *simple) help() string {
	return s.helpVal
}
func (s *simple) flags() *flag.FlagSet {
	return nil
}
