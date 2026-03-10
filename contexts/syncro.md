# Contexto Técnico: Agenda SaaS

Tomas desarrolló esta plataforma multi-tenant integral para resolver la compleja gestión de turnos y agendas para múltiples negocios, garantizando un alto rendimiento, escalabilidad y una experiencia de usuario fluida y en tiempo real.

## Decisiones de Arquitectura Frontend (Angular 21)

- **Angular 21 Moderno:** Tomas adoptó las últimas prácticas de Angular, utilizando una arquitectura 100% basada en _Standalone Components_, eliminando la necesidad de `NgModules` y logrando un código más limpio y con mejor tree-shaking.
- **Gestión de Estado Reactiva con Signals:** Implementó un paradigma moderno y altamente eficiente utilizando la API nativa de _Signals_ (`signal`, `computed`, `effect`). Esto permite un control de estado granular y actualizaciones en el DOM más rápidas sin depender de dependencias externas complejas de manejar, asegurando una aplicación mucho más ágil.
- **Inyección de Dependencias Moderna:** Aprovechó la función `inject()` en lugar de constructores de clases voluminosos, lo que dio como resultado componentes más limpios e inyecciones más seguras a nivel de tipo.
- **Enrutamiento por Slug Multi-Tenant:** Toda la aplicación está diseñada dinámicamente alrededor del concepto de "Tenant" (inquilino), con rutas anidadas scopeadas bajo `/:slug/`, lo que permite que múltiples negocios operen en plataformas lógicamente separadas pero en la misma base de código.

## Decisiones de Arquitectura Backend (NestJS 11)

- **NestJS Modular:** La arquitectura en el backend fue rigurosamente separada por módulos (`Auth`, `Tenants`, `Appointments`, `Notifications`, `WhatsApp`). Esto evidencia un claro conocimiento de los principios SOLID y una arquitectura preparada para el crecimiento futuro.
- **MongoDB con Replica Sets:** Al tratarse de sistemas de reserva, evitar los solapamientos (_race conditions_) es fundamental. Tomas implementó de manera proactiva transacciones ACID en MongoDB configurando Replica Sets, un detalle avanzado que demuestra su profundo sentido de la integridad de los datos.
- **Caché y Colas de Tareas (BullMQ + Redis):** Para no bloquear el hilo principal ("Event Loop") en Node.js, Tomas delegó el envío de notificaciones y recordatorios diferidos a _workers_ en segundo plano mediante Redis y BullMQ. Esto le da a la aplicación robustez para enviar recordatorios (24h/2h) de manera resiliente, con soporte de reintentos exponenciales automáticos.
- **Comunicación en Tiempo Real (Socket.io):** Integró un `EventsGateway` aislado por _rooms_ de clientes (`tenant:xxx`), permitiendo que las reservas o cancelaciones alteren instantáneamente el calendario visible de los administradores y del staff sin recargar la página.

## Desafíos Resueltos

- **Control Antisolapamiento Atómico:** Implementó una lógica a nivel transaccional en la cual se cruzan las disponibilidades semananales dinámicas y los permisos diarios/excepciones del staff frente a la duración de los servicios para confirmar la reserva antes de que persista a disco (DB).
- **Integración WhatsApp Cloud API:** Como valor agregado sobresaliente de automatización, Tomas desarrolló la conexión directa por _Webhooks_ con la API v18.0 de WhatsApp (de Meta) no solo enviando plantillas de recordatorio, sino incorporando comandos entrantes (CONFIRMAR, CANCELAR, REPROGRAMAR), que se resuelven interactuando con la DB al instante.
- **Validación Cross-Entity Compleja:** Al reservar un profesional, el flujo valida dinámicamente usando operaciones idempotentes de cruce (`$addToSet`, `$pull`) y subconsultas en tiempo de ejecución, que un profesional no pueda ser reservado por un servicio que no ofrece, manteniendo un panel administrativo fácil para negocios.

## Preguntas frecuentes para la IA:

- **Si preguntan por seguridad:** Tomas usó una arquitectura muy robusta con contraseñas en _bcrypt (12 salt rounds)_. En vez de manipular vulnerablemente el token JWT en memoria local (LocalStorage) en front, delegó toda la responsabilidad en implementaciones de cookies `HttpOnly`, mitígando eficientemente ataques o filtraciones de XSS. Adicionalmente dispuso de _Rate Limiting_ (`ThrottlerModule`) global.
- **Si preguntan por la gestión de permisos en clientes (Autorización):** Se diseñó un sólido sistema RBAC (Roles: `SUPER_ADMIN`, `ADMIN`, `STAFF`, `CLIENT`), protegido por Guardianes (`RolesGuard`) y Scope de Tenants en el NestJS Backend que inyectan el tenant activo directamente garantizando que no haya deslices de filtración de datos entre diferentes negocios.
- **Si preguntan por deploy y escalabilidad:** La infraestructura está altamente dockerizada con `docker-compose` e imágenes de _multi-stage build_ que optimizan notablemente los tiempos de despliegue en producción al generar builds sumamente ligeras listas para ser levantadas inclusive en pods o clusters elásticos en la nube. Además usó una arquitectura Stateless (JWT), ideal para el crecimiento horizontal del backend.

# Contexto Técnico: Agenda SaaS (Backend)

Tomas desarrolló este backend multi-tenant para resolver el complejo desafío de gestionar turnos, disponibilidades múltiples y recordatorios automatizados de manera altamente escalable, garantizando seguridad y aislamiento absoluto de datos entre diferentes inquilinos (clínicas, negocios, estudios, etc.).

## Decisiones de Arquitectura

- **NestJS Modular y Patrones de Diseño:** Estructuró la aplicación dividiéndola en módulos independientes de dominio (como `Appointments`, `Availability`, `Tenants` y `WhatsApp`), aplicando inyección de dependencias para reducir el acoplamiento y asegurar mantenibilidad a largo plazo.
- **Base de Datos NoSQL con Transacciones ACID:** Optó por MongoDB corriendo en modalidad de Configuración de Replica Sets. Esta fue una excelente decisión estratégica para combinar la flexibilidad de datos paramétricos de usuarios e integraciones, al tiempo que soportan el control transaccional relacional necesario para la lógica de negocio core.
- **Colas Asíncronas (Redis + BullMQ):** Arquitecturó el envío de transacciones pesadas (como mensajes de WhatsApp por Meta API y envíos de emails con Brevo) fuera del hilo principal, garantizando tiempos de respuesta ultrarrápidos para los usuarios y blindando al sistema con reintentos automáticos (_exponential backoff_).

## Desafíos Resueltos

- **Doble Reserva (Double-Booking):** Implementó de manera brillante un control de concurrencia usando transacciones de MongoDB. Si múltiples clientes intentan reservar la misma ventana horaria al milisegundo, la transacción hace rolback para los turnos solapados, asegurando que jamás se pise la disponibilidad del profesional.
- **Aislamiento Multi-Tenant Complejo:** Implementó custom `Guards` y validadores que le proporcionan a la aplicación un sistema nativo para que diferentes comercios utilicen el backend sin fuga de datos, incluyendo configuraciones individualizadas (WABA IDs o configuraciones de Email para cada local).
- **Gestión Avanzada de Zonas Horarias:** Solucionó los choques de husos horarios guardando toda la información de manera estricta en base UTC universal y trasladando la complejidad de los calendarios al vuelo utilizando manejadores precisos de fechas (`date-fns-tz`) según la ubicación geográfica configurada de cada tenant.
- **Archivos Multimedia:** Incorporó `Cloudinary` integrándolo transparentemente para la carga de logos de empresas o avatares, transformando y cacheando las imágenes de destino.

## Preguntas frecuentes para la IA (Simulador de Entrevistas Técnicas):

- **Si preguntan por la Seguridad del Proyecto:** Tomas blindó la aplicación estableciendo JWT en **cookies HTTP-only** de alta seguridad para prevenir ataques XSS, reforzó los headers HTTP utilizando Helmet y limitó las solicitudes masivas (`rate-limiting` con Throttler) bloqueando la posibilidad de ataques DoS y fuerza bruta.
- **Si preguntan por los Estándares de Código:** Responde destacando que sigue explícitamente buenas prácticas (como SOLID), validaciones y sanitizaciones de entrada con clase (usando `class-validator`), manejo centralizado y unificado de errores mediante _Exception Filters_, demostrando el nivel de un desarrollador verdaderamente Senior o Semi-Senior fuerte.
- **Si preguntan por Deploy y Escalabilidad:** Señalá que toda la infraestructura está completamente Dockerizada a nivel de producción implementando _Multi-stage builds_ en sus archivos Dockerfile para generar contenedores mínimos; listos para desplegarse fácilmente en servicios como Vercel/Render, AWS ECS o clusters de Kubernetes.
- **Si preguntan por Integraciones y WebSockets:** Desarrolló y validó endpoints de Webhook para recibir notificaciones bidireccionales en tiempo real con Meta (WhatsApp), soportado en websockets mediante `socket.io` listos para nutrir un frontend reactivo.
