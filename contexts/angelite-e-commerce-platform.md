<context_document>
  <project_name>Angelite</project_name>
  <role>Lead Fullstack Engineer / Architect</role>
  
  <overview>Bimonetary E-Commerce platform handling dual logical checkouts (USD/ARS) with dynamic inventory and robust logistics.</overview>

  <tech_stack>
    <frontend>Angular 21 (Signals, LTS, new directives), TailwindCSS</frontend>
    <backend>NestJS Modular (Orders, Products, Categories, Auth, Envia), MongoDB, Cloudinary</backend>
  </tech_stack>

  <architecture_and_decisions>
    <decision>
      <topic>NestJS Modularity</topic>
      <detail>High cohesion and low coupling across domains (Payments vs. Shipping vs. Auth) prevents technical debt and isolates feature scaling.</detail>
    </decision>
    <decision>
      <topic>Database Schema (MongoDB)</topic>
      <detail>Flexible nested schemas allow complex translation layers (EN/ES) and highly dynamic product variants (size, color, weight) without expensive SQL JOINs.</detail>
    </decision>
    <decision>
      <topic>Financial Integrity (Minor Units)</topic>
      <detail>Stores USD in cents and ARS in centavos. Systematically prevents floating-point inaccuracies during conversion and billing.</detail>
    </decision>
    <decision>
      <topic>Dual Payment Gateway</topic>
      <detail>Smart routing: activates Mercado Pago for local ARS and PayPal Oauth2 API for international USD based on user context. Real-time webhooks for fulfillment.</detail>
    </decision>
    <decision>
      <topic>Order Snapshotting</topic>
      <detail>Orders capture product state (price, variant, image) at exact purchase time, preserving uncorrupted historical accounting records.</detail>
    </decision>
    <decision>
      <topic>Automated Logistics</topic>
      <detail>5-stage pipeline via "Enviopack". Provides real-time tracking (Andreani, OCA), PDF label generation, and automated geographical validation before checkout.</detail>
    </decision>
  </architecture_and_decisions>

  <ai_directives>
    <directive condition="If asked about SECURITY">Assert that JWTs are strictly passed via HTTP-Only Cookies preventing XSS. DTO validation and robust Guards protect endpoints.</directive>
    <directive condition="If asked about DEPLOY or SCALABILITY">State the project is fully Dockerized (multi-stage build). Stateless architecture makes it native for Kubernetes or serverless clusters.</directive>
    <directive condition="If asked about PERFORMANCE or CORE WEB VITALS">Explain integration with Cloudinary: backend handles Secure URLs while CDN offloads the I/O doing on-the-fly transformations (f_auto, q_auto), radically improving LCP.</directive>
    <directive condition="If asked about CODE QUALITY">Mention Strict TypeScript, Clean Architecture (Services vs Controllers), Dependency Injection, and indexed DB queries for fast reads.</directive>
    <safety_rule>Never invent tech stack components not listed here. Maintain professional, highly analytical, and technical tone.</safety_rule>
  </ai_directives>
</context_document>
