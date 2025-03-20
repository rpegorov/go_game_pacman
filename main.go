package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

// Константы для типов объектов
const (
	ObjectPoop = iota
	ObjectBone
	ObjectCat
	ObjectStick
	ObjectTypesCount
)

// Константы для отображения
const (
	PlayerXPosition = 10 // Фиксированная позиция игрока по X
)

// GameConfig содержит настройки игры
type GameConfig struct {
	PlayerY           int           // Позиция игрока по Y
	InitialLives      int           // Начальное количество жизней
	BaseObjectSpeed   int           // Базовая скорость объектов
	FrameRate         time.Duration // Частота обновления кадров
	BaseSpawnRate     int           // Базовая частота появления объектов
	ScreenWidth       int           // Ширина экрана
	MaxSpeed          int           // Максимальная скорость объектов
	MinSpawnRate      int           // Минимальная частота появления объектов
	MouthOpenDuration int           // Продолжительность открытого рта
}

// DefaultGameConfig возвращает конфигурацию игры по умолчанию
func DefaultGameConfig() GameConfig {
	return GameConfig{
		PlayerY:           10,
		InitialLives:      5,
		BaseObjectSpeed:   1,
		FrameRate:         32 * time.Millisecond,
		BaseSpawnRate:     25,
		ScreenWidth:       80,
		MaxSpeed:          5,
		MinSpawnRate:      10,
		MouthOpenDuration: 8,
	}
}

// Sprite представляет графическое изображение объекта
type Sprite []string

// Object представляет игровой объект
type Object struct {
	X    int
	Type int
}

// LevelConfig содержит настройки сложности текущего уровня
type LevelConfig struct {
	ObjectSpeed int
	SpawnRate   int
}

// Game содержит игровое состояние
type Game struct {
	Config     GameConfig
	MouthOpen  bool
	MouthTime  int
	Lives      int
	Objects    []Object
	FrameCount int
	Score      int
	Sprites    GameSprites
}

// GameSprites содержит все спрайты игры
type GameSprites struct {
	DogClosed Sprite
	DogOpen   Sprite
	Objects   []Sprite
}

// NewGameSprites создает все спрайты для игры
func NewGameSprites() GameSprites {
	return GameSprites{
		DogClosed: Sprite{
			" \\___/ ",
			" (o o) ",
			" (___) ",
		},
		DogOpen: Sprite{
			" \\___/ ",
			" (o o) ",
			" (___O ",
		},
		Objects: []Sprite{
			// Какашка
			{
				" @@ ",
				"@@@@",
			},
			// Кость
			{
				"   ___",
				"__/   \\",
			},
			// Кот
			{
				" /\\_/\\ ",
				"( o.o )",
				" > ^ < ",
			},
			// Палка
			{
				"  _____",
				"_/     ",
			},
		},
	}
}

// GetObjectColor возвращает цвет для объекта заданного типа
func GetObjectColor(objectType int) termbox.Attribute {
	switch objectType {
	case ObjectPoop:
		return termbox.ColorMagenta
	case ObjectBone:
		return termbox.ColorWhite
	case ObjectCat:
		return termbox.ColorRed
	case ObjectStick:
		return termbox.ColorGreen
	default:
		return termbox.ColorDefault
	}
}

// ShowInstructions отображает инструкции к игре перед её началом
func ShowInstructions() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	width, height := termbox.Size()

	title := "DOG PACMAN - ИНСТРУКЦИЯ"
	DrawText(width/2-len(title)/2, 3, title, termbox.ColorYellow, termbox.ColorDefault)

	instructions := []string{
		"Ваша цель - управлять собакой, которая должна есть только определенные объекты.",
		"",
		"ОБЪЕКТЫ:",
		"  * КАКАШКИ (фиолетовые) - ЕДИМ! Открывайте рот, чтобы их съесть.",
		"  * КОСТИ (белые) - НЕ ЕДИМ! Держите рот закрытым.",
		"  * КОТЫ (красные) - НЕ ЕДИМ! Держите рот закрытым.",
		"  * ПАЛКИ (зеленые) - НЕ ЕДИМ! Держите рот закрытым.",
		"",
		"УПРАВЛЕНИЕ:",
		"  * ПРОБЕЛ - открыть рот",
		"  * ESC/Q - выйти из игры",
		"",
		"ПРАВИЛА:",
		"  * +1 очко за съеденную какашку",
		"  * -1 жизнь если съели не какашку",
		"  * -1 жизнь если пропустили какашку",
		"",
		"Нажмите любую клавишу для начала игры...",
	}

	startY := height/2 - len(instructions)/2
	for i, line := range instructions {
		DrawText(width/2-len(line)/2, startY+i, line, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
	termbox.PollEvent() // Ожидание нажатия клавиши
}

// NewGame создаёт новую игру с указанной конфигурацией
func NewGame(config GameConfig) *Game {
	return &Game{
		Config:     config,
		MouthOpen:  false,
		MouthTime:  0,
		Lives:      config.InitialLives,
		Objects:    []Object{},
		FrameCount: 0,
		Score:      0,
		Sprites:    NewGameSprites(),
	}
}

// GetLevelConfig возвращает настройки сложности на основе текущего счета
func (g *Game) GetLevelConfig() LevelConfig {
	// Базовые значения
	speed := g.Config.BaseObjectSpeed
	spawnRate := g.Config.BaseSpawnRate

	// Увеличиваем сложность каждые 10 очков
	levelIncrease := g.Score / 10

	// Увеличиваем скорость (до максимума)
	speed += levelIncrease
	if speed > g.Config.MaxSpeed {
		speed = g.Config.MaxSpeed
	}

	// Уменьшаем частоту появления объектов (делаем их чаще)
	spawnRate -= levelIncrease
	if spawnRate < g.Config.MinSpawnRate {
		spawnRate = g.Config.MinSpawnRate
	}

	return LevelConfig{
		ObjectSpeed: speed,
		SpawnRate:   spawnRate,
	}
}

// DrawText выводит текст на заданной позиции
func DrawText(x, y int, msg string, fg, bg termbox.Attribute) {
	for i, ch := range msg {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

// DrawSprite выводит многострочный спрайт на экран
func DrawSprite(x, y int, sprite Sprite, fg, bg termbox.Attribute) {
	for dy, line := range sprite {
		for dx, ch := range line {
			termbox.SetCell(x+dx, y+dy, rune(ch), fg, bg)
		}
	}
}

// Render отрисовывает текущее состояние игры
func (g *Game) Render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Отрисовка игрока (собаки)
	g.renderPlayer()

	// Отрисовка объектов
	g.renderObjects()

	// Отрисовка интерфейса
	g.renderUI()

	termbox.Flush()
}

// renderPlayer отрисовывает игрока (собаку)
func (g *Game) renderPlayer() {
	playerSprite := g.Sprites.DogClosed
	if g.MouthOpen {
		playerSprite = g.Sprites.DogOpen
	}
	DrawSprite(PlayerXPosition, g.Config.PlayerY, playerSprite, termbox.ColorYellow, termbox.ColorDefault)
}

// renderObjects отрисовывает все игровые объекты
func (g *Game) renderObjects() {
	for _, o := range g.Objects {
		model := g.Sprites.Objects[o.Type]
		color := GetObjectColor(o.Type)
		DrawSprite(o.X, g.Config.PlayerY, model, color, termbox.ColorDefault)
	}
}

// renderUI отрисовывает пользовательский интерфейс
func (g *Game) renderUI() {
	// Отрисовка информации о текущей игре
	levelConfig := g.GetLevelConfig()
	info := fmt.Sprintf("Lives: %d | Score: %d | Speed: %d", g.Lives, g.Score, levelConfig.ObjectSpeed)
	DrawText(0, 0, info, termbox.ColorWhite, termbox.ColorDefault)

	// Отрисовка инструкций
	instructions := "Space: Open mouth | ESC/Q: Quit"
	DrawText(g.Config.ScreenWidth-len(instructions), 0, instructions, termbox.ColorWhite, termbox.ColorDefault)

	// Отрисовка подсказки
	hint := "Eat only POOP, avoid other objects!"
	DrawText(g.Config.ScreenWidth/2-len(hint)/2, 2, hint, termbox.ColorCyan, termbox.ColorDefault)
}

// Update обновляет состояние игры
func (g *Game) Update() {
	// Обновляем состояние рта
	g.updateMouth()

	// Обновляем объекты
	g.updateObjects()

	// Создаем новые объекты
	g.maybeSpawnObject()

	g.FrameCount++
}

// updateMouth обновляет состояние рта собаки
func (g *Game) updateMouth() {
	if g.MouthOpen {
		g.MouthTime++
		if g.MouthTime >= g.Config.MouthOpenDuration {
			g.MouthOpen = false
			g.MouthTime = 0
		}
	}
}

// updateObjects обновляет положение и состояние всех объектов
func (g *Game) updateObjects() {
	levelConfig := g.GetLevelConfig()
	newObjects := []Object{}

	for _, o := range g.Objects {
		// Двигаем объект
		o.X -= levelConfig.ObjectSpeed

		// Проверяем столкновения
		if g.checkObjectCollision(o) {
			// Объект столкнулся с игроком и обработан
		} else if o.X > -len(g.Sprites.Objects[o.Type][0]) {
			// Объект еще на экране, сохраняем его
			newObjects = append(newObjects, o)
		}
	}

	g.Objects = newObjects
}

// checkObjectCollision проверяет и обрабатывает столкновение объекта с игроком
func (g *Game) checkObjectCollision(o Object) bool {
	objectWidth := len(g.Sprites.Objects[o.Type][0])
	objectHeight := len(g.Sprites.Objects[o.Type])
	dogWidth := len(g.Sprites.DogClosed[0])
	dogHeight := len(g.Sprites.DogClosed)

	// Проверка коллизии
	if CheckCollision(
		PlayerXPosition, g.Config.PlayerY, dogWidth, dogHeight,
		o.X, g.Config.PlayerY, objectWidth, objectHeight,
	) {
		// Обрабатываем столкновение по правилам игры
		g.handleCollision(o.Type)
		return true
	}

	return false
}

// handleCollision обрабатывает столкновение с объектом по правилам игры
func (g *Game) handleCollision(objectType int) {
	// Если рот открыт и это какашка - едим
	if g.MouthOpen && objectType == ObjectPoop {
		g.Score++
	} else if g.MouthOpen && objectType != ObjectPoop {
		// Рот открыт, но съели не какашку
		g.Lives--
	} else if !g.MouthOpen && objectType == ObjectPoop {
		// Рот закрыт, пропустили какашку
		g.Lives--
	}
	// Если рот закрыт и это не какашка - все нормально, ничего не делаем
}

// maybeSpawnObject создаёт новый объект с определённой вероятностью
func (g *Game) maybeSpawnObject() {
	levelConfig := g.GetLevelConfig()

	// Создание новых объектов с определенной вероятностью
	if g.FrameCount%levelConfig.SpawnRate == 0 && rand.Intn(3) > 0 {
		objectType := rand.Intn(ObjectTypesCount)
		g.Objects = append(g.Objects, Object{
			X:    g.Config.ScreenWidth,
			Type: objectType,
		})
	}
}

// CheckCollision проверяет столкновение двух прямоугольников
func CheckCollision(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// HandleInput обрабатывает пользовательский ввод
func (g *Game) HandleInput(ev termbox.Event) bool {
	if ev.Type == termbox.EventKey {
		switch {
		case ev.Key == termbox.KeyEsc || ev.Ch == 'q':
			return false
		case ev.Key == termbox.KeySpace && !g.MouthOpen:
			g.MouthOpen = true
			g.MouthTime = 0
		}
	}
	return true
}

// DrawGameOver отображает экран окончания игры
func DrawGameOver(score int) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	width, height := termbox.Size()

	// Графика "GAME OVER"
	gameOverArt := []string{
		"  ___    _    __  __   ___    _____   _   _  ___  ___ ",
		" / __|  /_\\  |  \\/  | | __|  / _ \\ \\ | | | |/ _ \\| _ \\",
		"| (_ | / _ \\ | |\\/| | | _|  |  __/\\ V /  | |  __/|   /",
		" \\___|/_/ \\_\\|_|  |_| |___|  \\___| \\_/   |_|\\___||_|_\\",
	}

	// Рисуем ASCII-арт
	artY := height/2 - len(gameOverArt) - 2
	for i, line := range gameOverArt {
		DrawText(width/2-len(line)/2, artY+i, line, termbox.ColorRed, termbox.ColorDefault)
	}

	// Сообщение о завершении игры
	finalScore := fmt.Sprintf("Final Score: %d", score)
	exitMsg := "Press any key to exit"

	DrawText(width/2-len(finalScore)/2, height/2+3, finalScore, termbox.ColorYellow, termbox.ColorDefault)
	DrawText(width/2-len(exitMsg)/2, height/2+5, exitMsg, termbox.ColorWhite, termbox.ColorDefault)

	termbox.Flush()

	// Ожидание нажатия клавиши
	termbox.PollEvent()
}

// RunGame запускает игровой цикл
func RunGame(game *Game) {
	// Создание игрового цикла
	gameLoop := time.NewTicker(game.Config.FrameRate)
	defer gameLoop.Stop()

	// Канал для событий пользовательского ввода
	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	// Главный цикл игры
	running := true
	for running && game.Lives > 0 {
		select {
		case <-gameLoop.C:
			game.Update()
			game.Render()
		case ev := <-eventQueue:
			running = game.HandleInput(ev)
		}
	}

	// Отображение экрана завершения игры
	DrawGameOver(game.Score)
}

// InitTermbox инициализирует termbox и настраивает его для игры
func InitTermbox() (int, error) {
	// Инициализация termbox
	err := termbox.Init()
	if err != nil {
		return 0, err
	}

	// Настройка событий
	termbox.SetInputMode(termbox.InputEsc)

	// Получение ширины экрана
	width, _ := termbox.Size()

	return width, nil
}

func main() {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Инициализация termbox
	width, err := InitTermbox()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	// Показываем инструкцию
	ShowInstructions()

	// Настройка конфигурации с учетом размера экрана
	config := DefaultGameConfig()
	config.ScreenWidth = width

	// Создание и запуск игры
	game := NewGame(config)
	RunGame(game)
}
