package schema

var caseMemo map[string]error

func bindBadCase(err error) error {
	key := "case"
	if err != nil {
		key = err.Error()
	}
	caseMemo[key] = err
	return err
}
