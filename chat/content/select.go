package content

// SelectItem представляет элемент списка.
type SelectItem struct {
	Caption string // Название.
	Data    string // Данные которые передаются в обработчик.
}

// Select представляет список отобржаемых элементов.
type Select struct {
	Header string       // Заголовок.
	Items  []SelectItem // Элементы списка.
}
