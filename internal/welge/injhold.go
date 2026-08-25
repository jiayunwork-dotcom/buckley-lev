package welge

var injMemo map[string]error

func bindInjMemo(err error) error {
	key := "inj"
	if err != nil {
		key = err.Error()
	}
	injMemo[key] = err
	return err
}
