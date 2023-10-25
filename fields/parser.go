package fields

type parser struct {
	sc        *scanner
	collector *fieldInfoCollector
}

func newParser(input string, collector *fieldInfoCollector) *parser {
	return &parser{
		sc:        newScanner(input),
		collector: collector,
	}
}

// =============================================
// Full Grammar
// =============================================
// FieldExpr => <Ident> FieldLevelList
// FieldLevelList => <Dot> <Ident> FieldLevelList
//				  | <Dot> FieldExprBracket
// 			      | <empty>
// FieldExprBracket => <Open Bracket> <FieldExpr> FieldSiblingList <Close Bracket>
// FieldSiblingList => <Vertical Line> <FieldExpr> FieldSiblingList
//					  | <empty>

func (p *parser) parse() error {
	if !p.sc.next() {
		return p.sc.withErrorf("missing field identifier")
	}
	err := p.parseFieldExpr(p.collector, parseFieldExprStateOutsideBracket)
	if err != nil {
		return err
	}
	if p.sc.getTokenType() != tokenTypeUnspecified {
		return p.sc.withErrorf("not allow extra string at the end")
	}
	return p.sc.getErr()
}

func (*parser) addParentPrefix(err error, prefix string) error {
	if err == nil {
		return err
	}
	if len(prefix) == 0 {
		return err
	}
	return PrependParentField(err, prefix)
}

type parseFieldExprState int

const (
	parseFieldExprStateOutsideBracket = iota + 1
	parseFieldExprStateStartOfBracket
	parseFieldExprStateMiddleOfBracket
)

func (p *parser) parseFieldExprGetErrorForFirstToken(state parseFieldExprState) error {
	if state == parseFieldExprStateOutsideBracket {
		return p.sc.withErrorf("expecting an identifier at the start, instead found '%s'", p.sc.getTokenString())
	}

	beforeToken := "|"
	if state == parseFieldExprStateStartOfBracket {
		beforeToken = "{"
	}
	return p.sc.withErrorf(
		"expecting an identifier after '%s', instead found '%s'",
		beforeToken, p.sc.getTokenString(),
	)
}

func (p *parser) parseFieldExprGetErrorForTokenIsNotDot(fieldElem string, state parseFieldExprState) error {
	if state == parseFieldExprStateOutsideBracket {
		if p.sc.getTokenType() != tokenTypeUnspecified {
			return p.sc.withErrorf(
				"expected '.' after identifier '%s', instead found '%s'",
				fieldElem,
				p.sc.getTokenString(),
			)
		}
	}
	return nil
}

//revive:disable-next-line:cognitive-complexity
func (p *parser) parseFieldExpr(coll *fieldInfoCollector, state parseFieldExprState) error {
	if p.sc.getTokenType() != tokenTypeIdent {
		return p.parseFieldExprGetErrorForFirstToken(state)
	}

	fieldElem := p.sc.getIdentString()
	parentPrefix := ""

	// FieldLevelList
	for {
		if !p.sc.next() || p.sc.getTokenType() != tokenTypeDot {
			if err := p.parseFieldExprGetErrorForTokenIsNotDot(fieldElem, state); err != nil {
				return err
			}
			return p.addParentPrefix(coll.addIfNotExisted(fieldElem, false), parentPrefix)
		}

		if err := coll.addIfNotExisted(fieldElem, true); err != nil {
			return p.addParentPrefix(err, parentPrefix)
		}

		var err error
		coll, err = coll.newSubCollector(fieldElem)
		if err != nil {
			return err
		}

		if len(parentPrefix) == 0 {
			parentPrefix = fieldElem
		} else {
			parentPrefix = fieldElem + "." + parentPrefix
		}

		if !p.sc.next() {
			return p.sc.withErrorf("expecting an identifier or a '{' after '.'")
		}

		if p.sc.getTokenType() == tokenTypeIdent {
			fieldElem = p.sc.getIdentString()
			continue
		}

		if p.sc.getTokenType() == tokenTypeOpeningBracket {
			return p.addParentPrefix(p.parseFieldExprBracket(coll), parentPrefix)
		}

		return p.sc.withErrorf(
			"expecting an identifier or a '{' after '.', instead found '%s'",
			p.sc.getTokenString(),
		)
	}
}

func (p *parser) parseFieldExprBracket(coll *fieldInfoCollector) error {
	if !p.sc.next() {
		return p.sc.withErrorf("expecting an identifier after '{'")
	}

	if err := p.parseFieldExpr(coll, parseFieldExprStateStartOfBracket); err != nil {
		return err
	}

	if err := p.parseFieldSiblingList(coll); err != nil {
		return err
	}

	if p.sc.getTokenType() != tokenTypeClosingBracket {
		if p.sc.getTokenType() == tokenTypeUnspecified {
			return p.sc.withErrorf("missing '}' at the end")
		}
		return p.sc.withErrorf("missing '}', instead found '%s'", p.sc.getTokenString())
	}

	p.sc.next()
	return p.sc.getErr()
}

func (p *parser) parseFieldSiblingList(coll *fieldInfoCollector) error {
	for {
		if p.sc.getTokenType() != tokenTypeVerticalLine {
			return nil
		}

		if !p.sc.next() {
			return p.sc.withErrorf("expecting an identifier after '|'")
		}

		if err := p.parseFieldExpr(coll, parseFieldExprStateMiddleOfBracket); err != nil {
			return err
		}
	}
}
