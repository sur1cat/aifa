# Atoma Development Workflow

## Как работать с агентами

### Шаг 1: Создай задачу
Создай файл в `/tasks/` с описанием фичи:

```bash
# Пример: tasks/feature-reminders.md
```

### Шаг 2: Product Manager
Открой Claude Code и напиши:
```
Прочитай agents/product-manager.md и tasks/[твоя-задача].md
Выступи как Product Manager и напиши спецификацию фичи.
Сохрани результат в specs/features/[название].md
```

### Шаг 3: Designer
```
Прочитай agents/designer.md и specs/features/[название].md
Выступи как Designer и спроектируй UI/UX.
Добавь результат в тот же файл specs/features/[название].md
```

### Шаг 4: Architect
```
Прочитай agents/architect.md и specs/features/[название].md
Выступи как Architect и спроектируй техническое решение.
Добавь результат в тот же файл specs/features/[название].md
```

### Шаг 5: Developer
```
Прочитай agents/developer.md и specs/features/[название].md
Выступи как Developer и реализуй фичу.
```

### Шаг 6: QA
```
Прочитай agents/qa.md и specs/features/[название].md
Выступи как QA и протестируй реализацию.
Создай чеклист в checklists/[название].md
```

### Шаг 7: DevOps
```
Прочитай agents/devops.md
Задеплой изменения на сервер.
```

---

## Быстрый старт (одной командой)

Для новой фичи напиши в Claude Code:

```
Новая фича: [описание]

1. Прочитай agents/product-manager.md, выступи как PM, напиши user story
2. Прочитай agents/designer.md, выступи как Designer, спроектируй UI
3. Прочитай agents/architect.md, выступи как Architect, спроектируй решение
4. Сохрани всё в specs/features/[название].md

Потом жди моего подтверждения перед разработкой.
```

После ревью:

```
Продолжи как Developer (agents/developer.md) — реализуй фичу по спецификации.
```

---

## Структура файлов

```
/habitflow
├── tasks/                    # Входящие задачи
│   └── feature-xxx.md
├── specs/
│   └── features/             # Спецификации фич (PM + Designer + Architect)
│       └── xxx.md
├── checklists/               # QA чеклисты
│   └── xxx.md
└── agents/                   # Промпты агентов
```
