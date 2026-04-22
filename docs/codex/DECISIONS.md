# DECISIONS.md

## 2026-04-23

- Настройки периодичности храним в колонке `tasks.periodicity_settings` типа `JSONB`.
- В Go используем typed-модель `task.PeriodicitySettings`, а не `map[string]any`.
- `nil` в `PeriodicitySettings` означает обычную непериодическую задачу.
- Режим "на конкретные даты" интерпретируем как список конкретных даты и времени.
- В доменной модели такие значения храним как `[]time.Time`, а в JSON/API и `JSONB` используем timestamp в формате RFC3339 с timezone, например `2026-10-25T10:00:00+03:00`.
