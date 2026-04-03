<context_document>
  <project_name>SKULLS Smash Burgers</project_name>
  <role>Fullstack Architect</role>
  
  <overview>Ultra-fast performant food-tech catalog and internal management portal driven by Engineering Pragmatism.</overview>

  <tech_stack>
    <frontend>Reactive logic (No explicit heavy framework mentioned, assume Vanilla/Lightweight JS), Minimalist UI</frontend>
    <backend>Express / Fastify, MongoDB (Atlas DaaS), Cloudinary</backend>
  </tech_stack>

  <architecture_and_decisions>
    <decision>
      <topic>MVP WhatsApp Checkout (Evolutionary Architecture)</topic>
      <detail>Delayed aggressive e-commerce cart implementation. Handles catalog and auth internally but delegates final checkout to a direct WhatsApp link. Faster time-to-market while setting ground for native `orders` collection.</detail>
    </decision>
    <decision>
      <topic>Embedded Documents Pattern (MongoDB)</topic>
      <detail>Used nested arrays for "Modifier Groups" (e.g. Extra Bacon, No Onions). Eliminates heavy SQL JOINS. Frontend reads massive reactive catalog trees in a single `find()` query without degrading DB performance.</detail>
    </decision>
    <decision>
      <topic>Strict Resource Separation</topic>
      <detail>Public read-only endpoints (`/api/products`) are heavily cached and physically decoupled from protected administrative mutating endpoints.</detail>
    </decision>
    <decision>
      <topic>Financial Typing (Nominal Mapping)</topic>
      <detail>Bypassed floating-point mathematical hazards by handling ARS currency exclusively via strict integer values across all layers.</detail>
    </decision>
    <decision>
      <topic>Cloudinary Processing</topic>
      <detail>Crucial for Gastronomy LCP (Largest Contentful Paint). Backend persists only Secure URLs; CDN automatically handles `f_auto`, `q_auto` to ensure mobile-first lightning speed.</detail>
    </decision>
  </architecture_and_decisions>

  <ai_directives>
    <directive condition="If asked about SECURITY">Auth requires JWTs (HTTP-only cookies) + encrypted passwords using `bcrypt`/`argon2`. Granular role Guards (`owner`, `editor`) restrict data mutation.</directive>
    <directive condition="If asked about SCALABILITY">Explicit use of Database Indexes (for category sorts, distinct slugs) ensures fast querying without bottlenecks. Architecture is purely Stateless, ready to be dockerized and scaled horizontally on Cloud Run/ECS.</directive>
    <directive condition="If asked to summarize the approach">Use the phrase "Engineering Pragmatism". Prioritized fast read speeds and intelligent DB modelling while intentionally avoiding premature engineering.</directive>
    <safety_rule>Keep responses concise. Emphasize product-oriented thinking. Focus on data indexing and time-to-market architecture.</safety_rule>
  </ai_directives>
</context_document>
