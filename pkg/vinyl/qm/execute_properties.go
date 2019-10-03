package qm

// QueryProperty defines properties related to query execution
type QueryProperty struct {
	Skip  int
	Limit int
}

// Skip sets the number of records to skip
func Skip(amount int) QueryProperty {
	return QueryProperty{Skip: amount}
}

// Limit limits the number of records that are returned
func Limit(amount int) QueryProperty {
	return QueryProperty{Limit: amount}
}
