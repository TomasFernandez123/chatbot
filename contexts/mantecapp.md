<context_document>
  <project_name>MantecApp</project_name>
  <role>Lead Technical Architect & Backend Developer</role>
  
  <overview>Comprehensive gastronomic management system (TIF 2025). Integrates real-time order tracking, complex RBAC, anonymous game UX, and async restaurant workflows.</overview>

  <tech_stack>
    <frontend>Angular 20 (Signals, Standalone), Ionic 8, Capacitor 7 (Native Hardware Integration)</frontend>
    <backend>Supabase (PostgreSQL, Auth, Storage, Realtime Node)</backend>
  </tech_stack>

  <architecture_and_decisions>
    <decision>
      <topic>Real-Time Sync (Kitchen/Bar Modules)</topic>
      <detail>Subscribed to Supabase Realtime via Angular Signals instead of polling. Maximizes hardware battery life for employee tablets and delivers zero-latency updates.</detail>
    </decision>
    <decision>
      <topic>Multilayer RBAC & Auth</topic>
      <detail>Implements multi-tiered JWT Guards supporting Employees, Owners, Registered Customers, and Anonymous Walk-ins (via QR). Social OAuth enabled.</detail>
    </decision>
    <decision>
      <topic>Async Approval Pipeline</topic>
      <detail>Walk-ins and reservations are placed in a SQL-queued holding state until explicitly approved/rejected via concurrent transactions by the owner.</detail>
    </decision>
    <decision>
      <topic>Hardware & Geo Integration</topic>
      <detail>Leveraged Capacitor 7 for QR barcode scanning (onboarding), haptics, and push notifications. Included leaflet/maplibre-gl for delivery GPS tracking.</detail>
    </decision>
    <decision>
      <topic>Performance Web Vitals (Frontend)</topic>
      <detail>Discarded Zone.js bottlenecks by using pure Signals for massive list renders (e.g., historical tickets). Extreme lazy-loading applied to game modules.</detail>
    </decision>
  </architecture_and_decisions>

  <ai_directives>
    <directive condition="If asked about SECURITY">Explain Row Level Security (RLS) in Supabase. Waitstaff cannot alter billing, and credentials are encrypted. JWTs are SSL-secured restricting CSRF and XSS vectors.</directive>
    <directive condition="If asked about LEADERSHIP or SOFT SKILLS">Highlight orchestration of a 3-dev team. Implemented Agile cycles, strictly enforced Single Responsibility Principle (SRP) via CI/CD linting, mitigating merge conflicts.</directive>
    <directive condition="If asked about DEPLOYMENTS">Frontend is compiled Ahead-of-Time (AOT) into a unified lightweight Android APK. Backend scales horizontally transparently via Supabase serverless infrastructure.</directive>
    <safety_rule>Focus explanations on the orchestration and architecture of real-time gastronomic workflows. Maintain an authoritative technical tone.</safety_rule>
  </ai_directives>
</context_document>
