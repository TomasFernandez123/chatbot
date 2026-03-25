# Chatbot (Go)

Backend HTTP en Go para consultas de chat con contexto por proyecto.

## Endpoint principal

`POST /chat`

El endpoint soporta dos modos:

1. **Clásico (buffered JSON)** — compatibilidad hacia atrás
2. **Streaming NDJSON** — tokens incrementales en tiempo real

---

## Modo clásico (backward compatible)

Si el cliente **no** envía `Accept: application/x-ndjson` (o envía `application/json`), el servidor responde como siempre con un JSON único:

```http
POST /chat
Accept: application/json
Content-Type: application/json

{"message":"Hola","project":"chatbot"}
```

Respuesta:

```json
{
  "answer": "...respuesta completa...",
  "timestamp": "2026-03-25T21:00:00Z"
}
```

---

## Modo streaming NDJSON

Para habilitar streaming, el cliente debe enviar:

- `Accept: application/x-ndjson`
- `Content-Type: application/json`

Ejemplo:

```bash
curl -N \
  -H "Accept: application/x-ndjson" \
  -H "Content-Type: application/json" \
  -X POST http://localhost:8080/chat \
  -d '{"message":"Explicame concurrencia en Go","project":"chatbot"}'
```

### Headers de respuesta (stream)

- `Content-Type: application/x-ndjson`
- `X-Accel-Buffering: off`
- `Cache-Control: no-cache`

### Frames NDJSON

Cada línea del stream es un JSON independiente terminado en `\n`.

#### Token frame

```json
{"type":"token","text":"fragmento parcial"}
```

#### Done frame (terminal exitoso)

```json
{"type":"done"}
```

#### Error frame (terminal de fallo)

```json
{"type":"error","message":"detalle del error"}
```

Notas:

- Si ocurre un error una vez iniciado el stream, el status HTTP puede permanecer en `200` y el cierre se comunica con frame `error`.
- El cliente debe procesar línea por línea (split por `\n`) y parsear cada frame JSON por separado.

---

## Compatibilidad hacia atrás

El streaming es **opt-in por header**. Clientes existentes que usen `application/json` (o sin `Accept`) mantienen el comportamiento clásico sin cambios.

---

## Proxy buffering (nginx/caddy)

Aunque el backend envía `X-Accel-Buffering: off`, el proxy también debe estar configurado para no bufferizar el stream.

### Nginx (ejemplo)

```nginx
location /chat {
    proxy_pass http://127.0.0.1:8080;

    # Requerido para streaming en tiempo real
    proxy_buffering off;
    proxy_request_buffering off;

    # Recomendado para conexiones largas
    proxy_read_timeout 180s;
    proxy_send_timeout 180s;

    # Respeta header de backend para desactivar buffering
    proxy_ignore_headers X-Accel-Buffering;
}
```

### Caddy (referencia rápida)

En Caddy v2, deshabilitá compresión/buffering para la ruta de streaming cuando sea necesario y mantené timeouts acordes al stream largo.
