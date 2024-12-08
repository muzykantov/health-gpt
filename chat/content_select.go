package chat

// SelectContentItem представляет элемент списка.
type SelectContentItem struct {
	Caption string // Название.
	Data    string // Данные которые передаются в обработчик.
}

// SelectContent представляет список отобржаемых элементов.
type SelectContent struct {
	Header string              // Заголовок.
	Items  []SelectContentItem // Элементы списка.
}
