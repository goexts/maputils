package extmap

type mapErr struct {
	v string
}

func (m mapErr) Error() string {
	return "map error:" + m.v
}
