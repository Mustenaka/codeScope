# codeScope Mobile — 打包 / 安装 / 测试指南

> 适用于无法直接使用 USB 安装（如小米设备 SIM 卡限制）的场景。  
> 所有操作均在项目根目录 `mobile/` 下执行。

---

## 1. 版本号规范

版本号统一在 `pubspec.yaml` 中维护：

```yaml
version: <major>.<minor>.<patch>+<build>
#        例如：  0.2.0+5
```

| 字段        | 含义                         | 示例   |
|-----------|------------------------------|--------|
| `major`   | 重大功能变更 / 破坏性改动      | `1`    |
| `minor`   | 新功能，向后兼容              | `2`    |
| `patch`   | Bug 修复                     | `3`    |
| `build`   | 构建流水号，每次出包 **+1**   | `42`   |

每次打包前，先更新 `pubspec.yaml` 中的版本号，版本信息会自动展示在 App 的**设置页 → 关于**中。

---

## 2. 环境检查

```powershell
flutter doctor        # 确认工具链无报错
flutter --version     # 确认 Flutter 版本
```

如刚修改过多语言文案（`l10n/*.arb`），建议先生成本地化代码：

```powershell
flutter gen-l10n
```

---

## 3. 打包 APK

### 3.1 Debug 包（内部测试，免签名）

```powershell
flutter build apk --debug
```

输出路径：
```
build/app/outputs/flutter-apk/app-debug.apk
```

### 3.2 Release 包（性能优化，需签名配置）

**首次配置签名**（仅做一次）：

```powershell
# 1. 生成 keystore
keytool -genkey -v -keystore keystore/codescope-release.jks -alias codescope_release -keyalg RSA -keysize 2048 -validity 10000

# 2. 复制配置模板
Copy-Item android\key.properties.example android\key.properties

# 3. 编辑 android/key.properties，填入真实密码和路径
```

**构建 Release 包：**

```powershell
flutter build apk --release
```

输出路径：
```
build/app/outputs/flutter-apk/app-release.apk
```

> Release 包体积更小、性能更好，推荐真机测试时使用。

---

## 4. 传输 APK 到手机

选择任意一种方式：

| 方式         | 说明                                 |
|------------|--------------------------------------|
| 微信 / QQ   | 发送给自己，手机端接收后点击安装       |
| 数据线文件传输 | 复制到手机存储，用文件管理器打开       |
| ADB（有 SIM 卡并已授权时） | `adb install build/.../app-debug.apk` |

---

## 5. 手机端安装

1. 打开手机文件管理器，找到 `app-debug.apk` / `app-release.apk`
2. 点击安装
3. 如提示"未知来源"：进入 **设置 → 安全 → 允许安装未知来源应用**，找到对应 App（文件管理器）并开启
4. 安装完成后，App 图标名称为 **codescope**

> 小米设备提示"需要插入 SIM 卡"时，仅影响 USB 直接安装，文件手动安装不受此限制。

---

## 6. 测试要点

### 6.1 确认版本

打开 App → **设置（右上角齿轮图标）**，在底部 **About / 关于** 区块中确认：

- **版本（Version）**：对应 `pubspec.yaml` 中 `major.minor.patch`
- **构建号（Build）**：对应 `pubspec.yaml` 中 `+build`

建议每次给你发测试包前按下面顺序操作：

```powershell
flutter pub get
flutter gen-l10n
flutter test
flutter build apk --debug
```

### 6.2 Mock 模式测试（默认）

| 功能             | 预期行为                                   |
|----------------|--------------------------------------------|
| 会话列表         | 展示 3 条 Mock 会话卡片                    |
| 进入会话详情     | 展示日志流，约 5 秒后 Mock 推送停止         |
| 文件浏览         | 占位页                                     |
| Prompt 下发     | 占位页                                     |
| 切换语言         | 设置页切换中文 / English，立即生效          |
| 切换数据源       | 切换为 Server，输入地址，保存后重新加载失败（服务端未实现） |

### 6.3 Server 模式测试（服务端就绪后）

1. 打开设置页
2. 数据源切换为 **Server**
3. 填入 REST API 地址和 WebSocket 地址
4. 点击 **保存**，返回首页验证真实数据加载

---

## 7. 本项目的注意事项（建议后续 agent / 开发者先看）

下面这些不是通用 Flutter 教程，而是本项目在实际修改和打包过程中已经遇到过的问题总结。

### 7.1 Flutter 版本差异不要想当然

- 本项目当前是按 **Flutter 3.41.x** 这一代 API 验证过的。
- 遇到 UI / Theme API 时，不要直接套用旧版本 Flutter 的写法。
- 已踩过的差异包括：
  - `ThemeData.cardTheme` 这里应使用 `CardThemeData`
  - `withOpacity()` 已被标记为不推荐，优先用 `withValues(alpha: ...)`
  - `surfaceVariant` 已被标记为不推荐，优先用 `surfaceContainerHighest`
- 结论：**遇到 Material 相关类型报错时，先确认当前 Flutter 版本对应的 API，再改代码**，不要一上来做大范围重构。

### 7.2 本项目的本地化不是“自动就会好”

- 本项目使用 `l10n.yaml` 管理本地化生成。
- 修改 `l10n/*.arb` 后，必须手动执行：

```powershell
flutter gen-l10n
```

- 当前生成产物输出在：

```text
lib/l10n/generated/
```

- 后续不要再假设一定能稳定使用 `package:flutter_gen/...` 的导入方式；本项目已经改为使用工程内生成文件的导入路径。
- 如果出现“找不到 `AppLocalizations`”或本地化导入异常，优先检查：
  1. 是否执行过 `flutter gen-l10n`
  2. `l10n.yaml` 是否仍然存在
  3. 代码是否仍引用到了旧的 `flutter_gen` 导入路径

### 7.3 版本号展示依赖运行时初始化

- App 版本号来自 `pubspec.yaml` 的：

```yaml
version: x.y.z+build
```

- 设置页里的版本 / 构建号展示由 `package_info_plus` 提供。
- 运行时必须先在 `main()` 中调用：
  - `WidgetsFlutterBinding.ensureInitialized()`
  - `AppVersion.initialize()`
- Widget test 环境不一定会走完整插件初始化流程，所以后续如果继续改版本逻辑，要注意：
  - 真机运行依赖初始化
  - 测试环境需要可回退、可容错，不能直接把页面渲染写死在插件初始化成功上

### 7.4 Windows / PowerShell 与 Bash 命令写法不同

- 这个仓库当前是在 **Windows + PowerShell** 环境下验证的。
- 命令链优先使用 `;`，不要默认写 Bash 风格的 `&&`。
- 复制文件优先使用：

```powershell
Copy-Item source target
```

- 文档里的命令应尽量保持 PowerShell 可直接复制执行，避免混入 Linux/macOS 专用写法。

### 7.5 Kotlin / Gradle 增量缓存异常不一定等于真正构建失败

- 在 Windows 下，加入 `package_info_plus` 后，曾遇到过：
  - `Invalid depfile`
  - Kotlin daemon / incremental cache 异常
- 这种情况下不要只看中间日志，要看最后是否出现：

```text
Built build\app\outputs\flutter-apk\app-debug.apk
```

- 如果最终 APK 已生成，则本次打包通常仍然是成功的。
- 如果确实失败，再按下面顺序清理：

```powershell
flutter clean
flutter pub get
flutter gen-l10n
flutter build apk --debug
```

- 如仍异常，再考虑删除项目下的 `build/`、`.dart_tool/` 后重试。

### 7.6 页面变长后，测试不能假设按钮一定在首屏

- 这次给设置页增加版本信息后，原有 widget test 因为“保存按钮不在可见区域”而失败。
- 结论：后续写 Flutter widget test 时，遇到 `ListView` / 可滚动页面，优先使用：
  - `scrollUntilVisible(...)`
  - 或先滚动再点击
- 不要默认页面新增内容后，原测试还能靠固定 `tap(find.text(...))` 稳定通过。

### 7.7 对安卓构建脚本只做最小改动

- 这次已经修过一次 `android/app/build.gradle.kts` 的兼容性问题。
- 后续如果不是明确的构建错误，不要因为“看起来是旧写法”就大改 Gradle / Kotlin DSL。
- 优先原则：
  1. 先确认是当前版本真实报错
  2. 再做最小修复
  3. 修完后立刻执行 `flutter analyze`、`flutter test`、`flutter build apk --debug` 验证

---

## 8. 常见问题

| 问题                          | 解决方案                                              |
|-----------------------------|-------------------------------------------------------|
| 构建失败 `Gradle error`      | 运行 `flutter clean; flutter pub get; flutter gen-l10n` 后重试 |
| 手机提示"解析包时出现问题"    | APK 可能传输损坏，重新传输后再安装                    |
| 安装后闪退                   | 先重新打包 debug APK 再测试；如需抓日志，可在设备已授权后再接 USB 排查 |
| 设置页看不到版本信息          | 确认 `AppVersion.initialize()` 在 `main()` 中已调用  |

