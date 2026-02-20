# Logging Policy

## Goals

- Keep logs structured and searchable.
- Make request tracing easy across services.
- Avoid logging sensitive data.

## Request Logging Standard

Request logs are emitted by middleware and include:

- `request_id`
- `method`
- `path`
- `status`
- `bytes`
- `duration_ms`
- `remote_ip`
- `user_agent`

## Log Levels

- `INFO`: successful requests (`2xx`, `3xx`)
- `WARN`: client errors (`4xx`)
- `ERROR`: server errors (`5xx`) and panics

## Sensitive Data Rules

Do not log:

- passwords
- tokens/API keys
- authorization headers
- full request/response bodies by default
- personal data unless strictly needed for debugging

## Error Logging Rules

- Return API errors via standard envelope (`code`, `message`).
- Internal errors can include details in logs, but API responses must stay generic.

## Future Improvements

- Add service-level structured logs in handlers/services.
- Add correlation fields for background jobs.
- Export logs to centralized storage/observability platform.
