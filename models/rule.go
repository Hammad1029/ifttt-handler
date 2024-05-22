package models

type RuleUDT struct {
	Operator1 ResolvableUDT   `cql:"op1" json:"op1"`
	Operand   string          `cql:"opnd" json:"opnd"`
	Operator2 ResolvableUDT   `cql:"op2" json:"op2"`
	Then      []ResolvableUDT `cql:"then" json:"then"`
	Else      []ResolvableUDT `cql:"else" json:"else"`
}

type QueryUDT struct {
	QueryString string          `cql:"query_str"`
	Resolvables []ResolvableUDT `cql:"resolvables"`
	Type        string          `cql:"type"`
}
