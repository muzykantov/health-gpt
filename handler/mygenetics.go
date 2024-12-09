package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/muzykantov/health-gpt/chat"
	"github.com/muzykantov/health-gpt/mygenetics"
	"github.com/muzykantov/health-gpt/server"
)

const MyGeneticsCodelabsPrompt = `
Ты - высококвалифицированный медицинский аналитик. Проанализируй предоставленные генетические данные и создай персонализированное заключение о состоянии здоровья, следуя этим шагам:
1. АНАЛИЗ ДАННЫХ
- Изучи все предоставленные генетические маркеры и их значения
- Определи взаимосвязи между различными генетическими показателями
- Оцени степень риска по каждому показателю
2. СТРУКТУРА ОТВЕТА
Предоставь анализ в следующем формате:
ОБЩЕЕ ЗАКЛЮЧЕНИЕ
- Краткое описание основных генетических особенностей
- Выявленные риски и их уровень (минимальный/умеренный/высокий)
- Общая оценка состояния здоровья
ДЕТАЛЬНЫЕ РЕКОМЕНДАЦИИ
Питание:
- Специфические диетические рекомендации с учетом генетического профиля
- Продукты, которые следует включить в рацион
- Продукты, которые следует ограничить
Образ жизни:
- Рекомендации по физической активности
- Рекомендации по режиму дня
- Профилактические меры
Медицинское наблюдение:
- Необходимые анализы и их периодичность
- Специалисты, консультации которых рекомендованы
- Параметры для мониторинга
3. ВАЖНЫЕ ПРАВИЛА
- Используй научно обоснованные рекомендации
- Учитывай взаимодействие различных генетических факторов
- Предоставляй конкретные, практические советы
- Укажи срочность выполнения рекомендаций
- Отметь, какие рекомендации требуют консультации с врачом
4. ФОРМАТ ВЫВОДА
🔬 Генетический профиль:
[Краткое описание основных генетических особенностей]
⚠️ Выявленные риски:
[Перечисление рисков с указанием уровня]
💡 Персональные рекомендации:
[Структурированные рекомендации по разделам]
⏰ Приоритетные действия:
[Что нужно сделать в первую очередь]
📋 Дополнительные замечания:
[Важные примечания и предостережения]
ПРАВИЛА ФОРМАТИРОВАНИЯ:
Используй только эмодзи в начале каждого раздела
Не используй markdown, жирный шрифт, курсив или другое сложное форматирование
Разделяй секции пустой строкой
Используй простые маркеры списка (•) для перечислений
Используй только простой текст
`

func MyGenetics() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			switch content := r.Incoming.Content.(type) {
			case string:
				myGeneticsCodelabs().Serve(ctx, w, r)

			case chat.SelectContentItem:
				myGeneticsCodelab(content.Data).Serve(ctx, w, r)

			default:
				w.WriteResponse(chat.NewMessage(chat.RoleAssistant,
					"⛔ Неизвестная команда. "+
						"Пожалуйста, выберите действие из предложенного списка."),
				)
			}
		},
	)
}

func myGeneticsCodelabs() server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			access := mygenetics.AccessToken(r.From.Tokens)

			codelabs, err := mygenetics.DefaultClient.FetchCodelabs(ctx, access)
			if err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						"⚠️ Не удалось загрузить список анализов. "+
							"Пожалуйста, попробуйте позже или обратитесь в поддержку.",
					),
				)
			}

			if len(codelabs) == 0 {
				w.WriteResponse(
					chat.NewMessage(chat.RoleAssistant,
						"⚠️ У вас пока нет доступных анализов. "+
							"Новые результаты появятся здесь автоматически."),
				)

				return
			}

			content := chat.SelectContent{
				Header: "🧪 Выберите анализ, чтобы просмотреть детальные результаты:",
			}
			for _, codelab := range codelabs {
				content.Items = append(content.Items, chat.SelectContentItem{
					Caption: fmt.Sprintf("%s (%s)", codelab.Name, codelab.Code),
					Data:    codelab.Code,
				})
			}

			w.WriteResponse(
				chat.NewMessage(chat.RoleAssistant, content),
			)

			content = chat.SelectContent{
				Header: "🧪 Выберите анализ для получения " +
					"развёрнутой интерпретации результатов с помощью ИИ:",
			}
			for _, codelab := range codelabs {
				content.Items = append(content.Items, chat.SelectContentItem{
					Caption: fmt.Sprintf("%s (%s)", codelab.Name, codelab.Code),
					Data:    "ai:" + codelab.Code,
				})
			}

			w.WriteResponse(
				chat.NewMessage(chat.RoleAssistant, content),
			)
		},
	)
}

func myGeneticsCodelab(code string) server.Handler {
	return server.HandlerFunc(
		func(ctx context.Context, w server.ResponseWriter, r *server.Request) {
			defer myGeneticsCodelabs().Serve(ctx, w, r)

			var useAI bool
			if strings.HasPrefix(code, "ai:") {
				code = strings.TrimPrefix(code, "ai:")
				useAI = true
			}

			w.WriteResponse(
				chat.NewMessage(
					chat.RoleAssistant,
					fmt.Sprintf("📡 Загружаю результаты анализа %s. "+
						"Это займёт несколько секунд...", code),
				),
			)

			features, err := mygenetics.DefaultClient.FetchFeatures(
				ctx,
				mygenetics.AccessToken(r.From.Tokens),
				code,
			)
			if err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						"⚠️ Не удалось получить информацию об анализе. "+
							"Пожалуйста, попробуйте позже или обратитесь в поддержку.",
					),
				)

				return
			}

			if !useAI {
				for i, feature := range features {
					time.Sleep(time.Millisecond * 300)
					select {
					case <-ctx.Done():
						return

					default:
						w.WriteResponse(
							chat.NewMessage(
								chat.RoleAssistant,
								feature.String()+
									"\n"+
									fmt.Sprintf(
										"📑 Показываю результат %d из %d.",
										i+1,
										len(features),
									),
							),
						)
					}
				}

				return
			}

			msgs := make([]chat.Message, 0, len(features)+1)

			msgs = append(msgs, chat.NewMessage(
				chat.RoleSystem,
				MyGeneticsCodelabsPrompt,
			))

			for _, feature := range features {
				msgs = append(msgs, chat.NewMessage(
					chat.RoleUser,
					feature.String(),
				))
			}

			w.WriteResponse(
				chat.NewMessage(
					chat.RoleAssistant,
					fmt.Sprintf("📑 Загружено %d параметров анализа. "+
						"Приступаю к обработке...", len(features)),
				),
			)

			w.WriteResponse(
				chat.NewMessage(
					chat.RoleAssistant,
					"⌛ Анализирую результаты с помощью ИИ. "+
						"Это может занять до минуты...",
				),
			)

			response, err := r.Completer.CompleteChat(ctx, msgs)
			if err != nil {
				w.WriteResponse(
					chat.NewMessage(
						chat.RoleAssistant,
						"⚠️ Не удалось получить интерпретацию результатов. "+
							"Пожалуйста, попробуйте позже или "+
							"просмотрите результаты без анализа ИИ.",
					),
				)

				return
			}

			w.WriteResponse(
				chat.NewMessage(
					chat.RoleAssistant,
					response.Content,
				),
			)
		},
	)
}
