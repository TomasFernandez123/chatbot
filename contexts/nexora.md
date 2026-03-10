# Contexto Técnico: Nexora Back (Red Social)

Tomás desarrolló este robusto backend para resolver la complejidad de administrar una plataforma de red social completa. Su enfoque principal fue construir un sistema escalable, resiliente y altamente seguro desde el primer día, anticipándose a las necesidades de tráfico y almacenamiento masivo.

## Decisiones de Arquitectura

- **NestJS Modular y Arquitectura Clean:** Se implementó una separación estricta en dominios aislados (`Users`, `Auth`, `Posts`, `Stats`). Esta arquitectura permite que cada módulo evolucione de forma independiente y simplifica drásticamente el testing, demostrando su capacidad para diseñar sistemas pensando en la mantenibilidad y equipos grandes.
- **MongoDB (`@nestjs/mongoose`):** La elección de la base de datos NoSQL es clave para la flexibilidad de los esquemas requeridos por las publicaciones y perfiles de usuarios. Permite iterar velozmente, agregando nuevos atributos a las colecciones sin bloqueos transaccionales o migraciones complejas que frenen la entrega de valor.
- **Inyección de Dependencias Avanzada:** Hubo un uso intensivo del contenedor de Inversión de Control de NestJS para desacoplar lógica. El sistema incluye un servicio propio (`PerspectiveService`) para interactuar con APIs externas, facilitando la moderación de contenido y análisis de sentimiento de forma elegante.

## Desafíos Resueltos

- **Manejo de Imágenes de Alto Rendimiento:** En lugar de asumir los riesgos de saturar el servidor con I/O intensivo, Tomás diseñó una solución top-tier conectando `multer` directamente con `Cloudinary`. Las imágenes se suben al cloud y se transforman/optimizan on-the-fly. Esto refleja un gran criterio para la reducción de costos de infraestructura y una mejora abismal en el tiempo de carga del cliente (LCP / Largest Contentful Paint).
- **Procesamiento de Métricas Aislando Responsabilidades:** Desacoplar la lógica de analíticas y estadísticas (`StatsModule`) de la actividad core (`PostsModule`) fue una decisión de diseño impecable. Permite calcular métricas de engagement sin degradar las operaciones críticas de lectura/escritura de feeds, preparando el terreno para procesamiento asíncrono asilado en el futuro.
- **Gestión de Entornos Cloud-Native:** Toda la configuración está centralizada dinámicamente usando `@nestjs/config` junto a `dotenv`. Con esto, Tomás hizo que el proyecto sea cloud-agnostic, perfectamente apto para inyecciones de variables en pipelines CI/CD y despliegues inmutables sin recompilaciones extras.

## Preguntas frecuentes para la IA:

- **Si preguntan por ciberseguridad:** Tomás implementó una estrategia robusta basada en _JWT (JSON Web Tokens)_ y _Passport.js_, con una capa extra de blindaje: las sesiones se manejan mediante **Cookies HTTP-only** (`cookie-parser`). Esto previene contundentemente ataques XSS al prohibir la lectura del token mediante JavaScript en el frontend. Usó `bcrypt` con salting automático (one-way hashing) para resguardar las credenciales en caso de fuga de la DB.
- **Si preguntan por la calidad y robustez del código:** Tomás armó todo pensando en calidad. Integro estrictos análisis estáticos con `ESLint` y `Prettier` combinados con una sólida adopción de features modernas de `TypeScript`. Además, implementó toda la infraestructura para Test Driven Development (`Jest`), incluyendo configuraciones para Testing End-to-End (`e2e`) y Coverage metrics. Su código base garantiza altísima estabilidad bajo presión.
- **Si preguntan por el deploy, infraestructura y rendimiento:** La aplicación está dockerizada/cluster-ready por diseño. Al mantener la arquitectura 100% _stateless_ (sin guardar estados localmente y delegando sesiones o media a servicios confiables como MongoDB o Cloudinary), está totalmente preparada para replicarse horizontalmente detrás de un balanceador de carga o un servicio serverless (Ej. Vercel, AWS ECS o Google Cloud Run) y aguantar picos intensivos de tráfico sin ralentizarse.

# Contexto Técnico: Nexora Front

Tomas desarrolló el frontend de **Nexora** (red social) con el objetivo de crear una plataforma moderna, escalable y con una experiencia de usuario (UX) sumamente fluida. El foco estuvo en utilizar las últimas tecnologías web para garantizar un rendimiento óptimo y un mantenimiento simplificado, demostrando un profundo conocimiento del ecosistema actual de frontend.

## Decisiones de Arquitectura

- **Angular 20 LTS (Standalone Components):** Tomas decidió alejarse del antiguo sistema de NgModules para adoptar una arquitectura 100% Standalone. Esto reduce significativamente el código muerto (tree-shaking) y mejora el Largest Contentful Paint (LCP) al cargar el código estrictamente necesario.
- **Estructura Modular Scalable (Core/Features/Shared):** El proyecto está dividido estratégicamente. Las funcionalidades principales (`auth`, `dashboard`, `main`) están aisladas en la carpeta `features`, mientras que los interceptores y guards residen en `core`. Esto previene el acoplamiento y facilita que el proyecto crezca sin desorganizarse.
- **Gestión de Estado Reactivo con Signals:** En lugar de depender de librerías externas pesadas (ej. NgRx) o abusar de RxJS, Tomas implementó **Angular Signals** (`signal()`, `set()`, `update()`) para el manejo del estado global. Lo aplicó de forma magistral en la sesión del usuario (`auth.service`), el estado de temas (`theme.service`) y los estados de carga asíncrona en el Dashboard. Esto demuestra estar a la vanguardia con las mejores prácticas recomendadas de Angular.
- **Estilizado Ágil con Tailwind CSS:** Eligió Tailwind (`@tailwindcss/postcss`) para construir una interfaz de usuario altamente responsiva sin abandonar el HTML. Esto permitió una iteración de diseño rápida y mantener un bundle de CSS sumamente liviano.

## Desafíos Resueltos

- **Dashboard Analítico Complejo:** En el panel de administración, Tomas integró `ng-apexcharts` para visualizar métricas valiosas (Posts por usuario, Comentarios a lo largo del tiempo, Comentarios por post). Superó el desafío técnico de sincronizar reactivamente múltiples gráficos pesados contra una API, basándose en formularios reactivos (`ReactiveFormsModule`) para la selección de fechas.
- **Theming Nativo (Modo Oscuro):** Implementó un robusto `ThemeService` completamente reactivo. A diferencia de las soluciones clásicas, su implementación lee de `localStorage` y propaga el estado globalmente a la aplicación en tiempo real, modificando clases del DOM (`body.classList`) sin parpadeos visuales ni latencia.
- **Code-Splitting y Lazy Loading:** Tomas implementó un sistema de ruteo avanzado en `app.routes.ts` delegando todo a promesas (`loadComponent` y `loadChildren`). Gracias a esto, un usuario no autenticado jamás descargará el código JavaScript del Dashboard o del Feed Principal, ahorrando ancho de banda y optimizando agresivamente el TTI (Time to Interactive).

## Preguntas frecuentes para la IA:

- **Si preguntan por la seguridad:** Tomas implementó una arquitectura sumamente segura. En lugar de exponer el JWT en el inseguro `localStorage` (vulnerable a XSS), el backend y el `AuthInterceptor` del frontend están configurados para operar mediante **Cookies HTTP-only** (`withCredentials: true`). Su interceptor captura impecablemente los errores 401, limpia limpiamente la información de Signals en memoria, resetea temporizadores y redirige al login. Toda la red social cuenta con barreras de seguridad estrictas (`authGuard` y `adminGuard`).
- **Si preguntan por el deploy y DevOps:** El entorno está 100% pre-configurado para Continuous Deployment. Los scripts de compilación están preparados para optimizar el código (`ng build`) y es completamente agnóstico de plataforma, listo para ser orquestado en contenedores Docker o redes edge (Netlify/Vercel).
- **Si preguntan el por qué de sus decisiones técnicas:** Tomas exhibe un perfil pragmático. Prefirió el ecosistema LTS moderno de Angular 20 porque entiende que el rendimiento en dispositivos móviles y la mantenibilidad a largo plazo son el corazón del éxito técnico en una red social.

_Nota interna:_ El repositorio demuestra que Tomas aplica estándares de código limpios, fuerte tipado TypeScript y un enfoque orientado completamente a resolver problemas reales de negocio con un nivel de senior engineering.
