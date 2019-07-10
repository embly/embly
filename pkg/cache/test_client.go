package cache

func NewTestClient() Client {
	tc := testClient{
		values: make(map[string]interface{}),
	}
	return &tc
}

type testClient struct {
	values map[string]interface{}
}

func (tc *testClient) Set(key, val string) error {
	tc.values[key] = val
	return nil
}
func (tc *testClient) Get(key string) (string, error) {
	switch v := tc.values[key].(type) {
	case string:
		return v, nil
	default:
		return "", nil
	}
}
func (tc *testClient) SetB(key string, val []byte) error {
	tc.values[key] = val
	return nil
}
func (tc *testClient) GetB(key string) ([]byte, error) {
	switch v := tc.values[key].(type) {
	case []byte:
		return v, nil
	default:
		return nil, nil
	}
}
