# Services Layer

This folder contains:

- REST API clients
- WebSocket client
- request models
- response models

Current structure:

- `mock/`
  - Mock REST and WebSocket implementations used when `AppEnvironment.useMock == true`
- `real/`
  - Real server REST and WebSocket implementations used when `AppEnvironment.useMock == false`
- `rest_client.dart`
  - Shared REST abstraction for controllers
- `websocket_client.dart`
  - Shared WebSocket abstraction for controllers
- `app_services.dart`
  - Composition root that swaps mock and real services without changing page code
- `connection_tester.dart`
  - Abstracts connection probing away from the Settings page
- `default_connection_tester.dart`
  - Verifies `/api/health` and `/ws/mobile` for the currently entered server settings
