package valse2

type UserValue struct {
	d map[string]interface{}
}

func (u UserValue) Set(k string, v interface{}) UserValue {
	u.d[k] = v
	return u
}

func (u UserValue) Get(k string) interface{} {
	return u.d[k]
}
