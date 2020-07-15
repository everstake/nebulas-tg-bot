package models

type Dictionary map[string]map[string]string

func (d Dictionary) Get(key string, lang string) string {
	mp, ok := d[key]
	if !ok {
		return ""
	}
	return mp[lang]
}
