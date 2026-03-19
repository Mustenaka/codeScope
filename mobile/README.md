# codeScope Mobile

Flutter MVP client for `codeScope`, focused on session list, session detail log stream, and extension points for file browsing and prompt dispatch.

Current UI also includes:

- Runtime language switch: Chinese / English
- Settings page with REST API / WebSocket connection fields
- Runtime-only environment switching between mock data and server mode
- Version / build display in Settings → About

## Current State

This implementation is still mock-first at startup, but the core Session and Log flows now support the real server APIs.

### Mocked parts

- `lib/services/mock/mock_rest_client.dart`
  - Simulates `GET /api/sessions`
  - Simulates `GET /api/sessions/:id`
  - Simulates `GET /api/sessions/:id/events`
- `lib/services/mock/mock_websocket_client.dart`
  - Simulates live log append
- `lib/services/mock/mock_data_provider.dart`
  - Provides seeded session and event samples

### Connected real APIs

- `lib/services/real/server_rest_client.dart`
  - Calls `GET /api/sessions`
  - Calls `GET /api/sessions/:id`
  - Calls `GET /api/sessions/:id/events`
- `lib/services/real/server_websocket_client.dart`
  - Connects to `GET /ws/mobile?session_id=<id>`
  - Decodes server `event.Record` payloads into `LogEvent`
- `lib/services/real/server_api_mapper.dart`
  - Maps server JSON fields into the existing mobile MVP models

### Runtime settings

- Language and connection settings are applied immediately in memory
- They are not persisted locally yet
- Real local persistence can be added later with `shared_preferences` or another storage layer
- App startup now reads optional Dart defines in `lib/main.dart`
- To switch to the real backend during a run:
  1. Open Settings
  2. Change data source to `Server`
  3. Enter the REST API base URL and WebSocket URL
  4. Use `Test connection` to probe `/api/health` and `/ws/mobile`
  5. Save settings and reload the session list

### Startup with real backend

You can start directly in server mode without opening Settings:

```powershell
flutter run --dart-define=CODESCOPE_API_BASE_URL=http://127.0.0.1:8080/api --dart-define=CODESCOPE_WS_URL=ws://127.0.0.1:8080/ws/mobile
```

Optional override:

```powershell
flutter run --dart-define=CODESCOPE_USE_MOCK=true
```

Behavior:

- If `CODESCOPE_USE_MOCK=true`, the app starts in mock mode
- If `CODESCOPE_USE_MOCK=false`, the app starts in server mode
- If `CODESCOPE_USE_MOCK` is omitted but either URL define is provided, the app starts in server mode
- If no defines are provided, the app still starts in mock mode

### Current integration boundaries

- `lib/services/rest_client.dart`
  - Declares the REST abstraction consumed by controllers
- `lib/services/websocket_client.dart`
  - Declares the WebSocket abstraction consumed by controllers
- `lib/services/app_services.dart`
  - Chooses mock vs real service implementations from `AppEnvironment.useMock`

Mock and real implementations both sit behind these abstractions, so pages and controllers do not know which transport is active.

## Page Structure

- `SessionPage`
  - Home page and session list
- `LogPage`
  - Session summary and live log timeline
- `FileBrowserPage`
  - Placeholder for workspace tree and content viewer
- `PromptPage`
  - Placeholder for prompt library and dispatch flow

## Suggested Real Server Integration

1. Keep controllers unchanged:
   - `SessionController` only depends on `CodeScopeRestClient`
   - `LogController` only depends on `CodeScopeRestClient` and `CodeScopeWebSocketClient`
2. Expand placeholders with future APIs:
   - Files: `/api/sessions/:id/files/tree`, `/content`
   - Prompt: `/api/prompts`, `/api/sessions/:id/commands/prompt`
3. Add reconnect, heartbeat-state UI, and persistence later without changing page-level code

## Local Development

```powershell
flutter pub get
flutter gen-l10n
flutter test
flutter analyze
```

If your Android device cannot install over USB directly (for example, some Xiaomi devices without a SIM card inserted), use manual APK packaging instead of `flutter run`.

See [`BUILDING.md`](BUILDING.md) for:

- version bumping rules
- debug / release APK packaging
- manual installation on phone
- smoke-test checklist

## Android Packaging Notes

- Package name: `com.codescope.mobile`
- App name: `codescope`
- Network permission is enabled in Android manifest for future API and WebSocket access.

### Temporary release signing

`android/app/build.gradle.kts` now uses this behavior:

- If `android/key.properties` exists, use the configured release keystore.
- If `android/key.properties` does not exist, fall back to debug signing for temporary release packaging.

This is suitable for internal testing builds, not store release.

### Release signing template

1. Copy `android/key.properties.example` to `android/key.properties`
2. Create a keystore file, for example at `mobile/keystore/codescope-release.jks`
3. Replace the placeholder values in `android/key.properties`

Example:

```properties
storePassword=your-store-password
keyPassword=your-key-password
keyAlias=codescope_release
storeFile=../keystore/codescope-release.jks
```

After that, Android Studio or Flutter release builds will automatically use the release signing config.
