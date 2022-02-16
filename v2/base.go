package extmap

func Merge(sources ...map[string]any) map[string]any {
	target := make(map[string]any)
	if sources == nil {
		return target
	}
	for _, v := range sources {
		for k, v := range v {
			target[k] = v
		}
	}
	return target
}
