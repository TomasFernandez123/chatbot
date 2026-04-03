<context_document>
  <project_name>Nexora</project_name>
  <role>Lead Fullstack Engineer</role>
  
  <overview>Robust, highly scalable Social Network platform featuring async metric processing, dynamic dashboards, and high-performance media delivery.</overview>

  <tech_stack>
    <frontend>Angular 20 (Standalone Components, Signals), TailwindCSS, ng-apexcharts</frontend>
    <backend>NestJS (Modular, Clean Architecture), MongoDB (@nestjs/mongoose), Cloudinary</backend>
  </tech_stack>

  <architecture_and_decisions>
    <decision>
      <topic>Standalone Angular Architecture</topic>
      <detail>Removed NgModules entirely. Drastically reduces dead code (tree-shaking) optimizing TTI and LCP compared to older Angular patterns.</detail>
    </decision>
    <decision>
      <topic>Reactive State without NgRx</topic>
      <detail>Utilized Angular Signals (signal, computed, effect) for global state (Auth, Theme) eliminating boilerplate and heavy 3rd-party dependencies.</detail>
    </decision>
    <decision>
      <topic>Dynamic Theming</topic>
      <detail>Native DOM interaction (body.classList) tied to LocalStorage with reactive updates preventing UI flicker.</detail>
    </decision>
    <decision>
      <topic>NestJS Modularity</topic>
      <detail>Strict separation between Users, Auth, Posts, and Stats boundaries simplifies E2E and unit testing.</detail>
    </decision>
    <decision>
      <topic>Async Stats Computation</topic>
      <detail>Isolating "StatsModule" from "PostsModule" ensures that heavy engagement metric computations do not degrade core CRUD I/O performance of the social feed.</detail>
    </decision>
    <decision>
      <topic>Cloudinary Media Pipeline</topic>
      <detail>Direct multer-to-Cloudinary uploading prevents intensive disk I/O on the application server. Delivers transformed images directly via CDN.</detail>
    </decision>
    <decision>
      <topic>Code-Splitting / Lazy Loading</topic>
      <detail>Heavy dashboard charts and core feeds are deferred via lazy `loadComponent` promises, saving massive bandwidth for unauthenticated users.</detail>
    </decision>
  </architecture_and_decisions>

  <ai_directives>
    <directive condition="If asked about SECURITY">Highlight JWT + Passport.js integrated via HTTP-only withCredentials Cookies. Bypasses XSS entirely. Bcrypt w/ automatic salting mitigates DB breaches.</directive>
    <directive condition="If asked about CODE QUALITY">Emphasize Test Driven Development (TDD) via Jest (E2E + Coverage), strict ESLint/Prettier CI pipelines, and advanced TypeScript features.</directive>
    <directive condition="If asked about INFRASTRUCTURE or DEPLOYMENTS">State the app is cloud-native (stateless). Utilizes @nestjs/config with dotenv. Fully Dockerized, cluster-ready for horizontal scaling under heavy traffic spikes.</directive>
    <safety_rule>Never hallucinate features or libraries (e.g. do not mention Redux, React, SQL). The focus is Modern Angular + NestJS Pragmatism.</safety_rule>
  </ai_directives>
</context_document>
