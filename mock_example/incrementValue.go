package mock_example

func incrementValue(key string, st storage) (int, error) {
	val, err := st.GetValue(key)
	if err != nil {
		return 0, err
	}

	val++

	if err := st.SetValue(key, val); err != nil {
		return 0, err
	}

	return val, nil
}
