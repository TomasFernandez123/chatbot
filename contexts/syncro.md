<context_document>
  <project_name>Agenda SaaS</project_name>
  <role>Lead Fullstack Architect</role>
  
  <overview>Multi-tenant scheduling and booking platform built for high-performance scale, avoiding race conditions and ensuring absolute data isolation across different business tenants.</overview>

  <tech_stack>
    <frontend>Angular 21 (Standalone Components, Signals, `inject()` paradigm)</frontend>
    <backend>NestJS 11 (Strict Modular), MongoDB (Replica Sets), Redis + BullMQ, Socket.io</backend>
  </tech_stack>

  <architecture_and_decisions>
    <decision>
      <topic>MongoDB Replica Sets (ACID Transactions)</topic>
      <detail>Enforced strict multi-document ACID transactions. Prevents Double-Booking race conditions if multiple users reserve the same exact millisecond slot, triggering automatic rollbacks.</detail>
    </decision>
    <decision>
      <topic>Async Task Queues (Redis/BullMQ)</topic>
      <detail>Offloaded heavy I/O operations (WhatsApp/Email reminders) to background workers. Protects the Node.js Event Loop from blocking and supports automated exponential backoff retries.</detail>
    </decision>
    <decision>
      <topic>Multi-Tenant Routing & Isolation</topic>
      <detail>Frontend structured by `/:slug/` scopes. Backend strictly uses tenant-injected Custom Guards ensuring complete data isolation so different clinics/businesses do not leak DB scopes.</detail>
    </decision>
    <decision>
      <topic>WhatsApp Cloud API Real-Time Interactivity</topic>
      <detail>Built bi-directional Webhooks via Meta v18 API handling incoming templates (CONFIRM, CANCEL) triggering DB state changes and emitting Socket.io events to update frontend dashboards instantly.</detail>
    </decision>
    <decision>
      <topic>Advanced Authorization (RBAC)</topic>
      <detail>Implemented `RolesGuard` logic (`SUPER_ADMIN`, `ADMIN`, `STAFF`, `CLIENT`) coupled with idempotent DB operators (`$addToSet`, `$pull`) validating cross-entity relations during bookings.</detail>
    </decision>
  </architecture_and_decisions>

  <ai_directives>
    <directive condition="If asked about SECURITY">Strict usage of HTTP-Only cookies for JWT preventing XSS. Implementation of Global Rate Limiting (`ThrottlerModule`) mitigates DoS and brute-force attacks.</directive>
    <directive condition="If asked about SCALABILITY or DEPLOYMENT">App is 100% Dockerized via multi-stage builds. Backend is fully Stateless making it suitable to scale elastically across AWS ECS or Kubernetes clusters.</directive>
    <directive condition="If asked about CODE QUALITY">Adheres to SOLID principles, payload sanitization (`class-validator`), and central Exception Filters indicating a clear Senior engineering mindset.</directive>
    <safety_rule>Never deviate from the multi-tenant nature of the app. Frame architectural decisions around "preventing race conditions" and "tenant data isolation".</safety_rule>
  </ai_directives>
</context_document>
