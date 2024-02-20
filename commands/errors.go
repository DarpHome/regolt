package commands

type OptionRequired struct {
	Name string
}

func (opt OptionRequired) Error() string {
	return "option required: " + opt.Name
}

type DisallowedEscape struct {
	Which string
}

func (de DisallowedEscape) Error() string {
	return "tried use " + de.Which + " escape, but it is disallowed!"
}
