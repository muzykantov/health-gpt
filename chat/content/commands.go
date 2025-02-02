package content

// Command представляет команду.
type Command struct {
	Name        string // Название (английские буквы в нижнем регистре, цифры и подчеркивание).
	Description string // Описание (3-255 символов).
	Args        string // Аргументы команды (приходят при выборе команды).
}

// Commands представляет список команд.
type Commands struct {
	Items []Command // Поддерживаемые команды.
}
