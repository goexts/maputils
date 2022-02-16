package extmap

type Setting struct {
	Split bool
}

func defaultSetting() *Setting {
	return &Setting{Split: true}
}

type SettingOption func(op *Setting)
