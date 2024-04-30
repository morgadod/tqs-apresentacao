package cache

type Cache struct {
	ids map[string][]int
}

func New() *Cache {
	return &Cache{
		ids: make(map[string][]int),
	}
}

func (cache *Cache) AddId(id int, objectName string) {
	if cache.IdExists(id, objectName) {
		return
	}

	cache.ids[objectName] = append(cache.ids[objectName], id)
}

func (cache *Cache) IdExists(id int, objectName string) bool {
	for _, existingID := range cache.ids[objectName] {
		if existingID == id {
			return true
		}
	}

	return false
}
