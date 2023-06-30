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
	return p.parseFieldExpr(p.collector)
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

func (p *parser) parseFieldExpr(coll *fieldInfoCollector) error {
	if p.sc.getTokenType() != tokenTypeIdent {
		return p.sc.withErrorf("expecting an identifier, instead found '%s'", p.sc.getTokenString())
	}

	fieldElem := p.sc.getIdentString()
	parentPrefix := ""

	// FieldLevelList
	for {
		if !p.sc.next() || p.sc.getTokenType() != tokenTypeDot {
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

	if err := p.parseFieldExpr(coll); err != nil {
		return err
	}

	if err := p.parseFieldSiblingList(coll); err != nil {
		return err
	}

	if p.sc.getTokenType() != tokenTypeClosingBracket {
		return p.sc.withErrorf("missing '}'")
	}
	p.sc.next()

	return nil
}

func (p *parser) parseFieldSiblingList(coll *fieldInfoCollector) error {
	for {
		if p.sc.getTokenType() != tokenTypeVerticalLine {
			return nil
		}

		if !p.sc.next() {
			return p.sc.withErrorf("expecting an identifier after '|'")
		}

		if err := p.parseFieldExpr(coll); err != nil {
			return err
		}
	}
}
