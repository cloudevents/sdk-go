package v2

type Parser struct {
	// TODO parser options
}

func (p *Parser) Parse(input string) (Expression, error) {
	return nil, nil
}

var defaultParser = Parser{}

func Parse(input string) (Expression, error) {
	return defaultParser.Parse(input)
}
