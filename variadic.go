package tablewriter

// SetHeaderVariadic takes an variadic argument, e.g. funcName("this", "is", "variadic")
func (t *Table) SetHeaderVariadic(keys ...string) {
	t.SetHeader(keys)
}

// SetFooterVariadic takes an variadic argument, e.g. funcName("this", "is", "variadic")
func (t *Table) SetFooterVariadic(keys ...string) {
	t.SetFooter(keys)
}

// AppendVariadic takes an variadic argument, e.g. funcName("this", "is", "variadic")
func (t *Table) AppendVariadic(row ...string) {
	t.Append(row)
}
