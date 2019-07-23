package eval

var order = []map[string]bool{
	{"%": true},
	{"+": true, "-": true},
	{"*": true, "&": true, "|": true},
}
