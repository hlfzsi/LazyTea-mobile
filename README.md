<div align="center">

<a href="https://github.com/hlfzsi/LazyTea-mobile">
  <img src="https://socialify.git.ci/hlfzsi/LazyTea-mobile/image?description=1&descriptionEditable=%E2%9C%A8%20%E4%B8%80%E6%AC%BE%E8%B7%A8%E5%B9%B3%E5%8F%B0%E7%A7%BB%E5%8A%A8%E5%AE%A2%E6%88%B7%E7%AB%AF%20%E2%9C%A8&issues=1&language=1&logo=https%3A%2F%2Fraw.githubusercontent.com%2Fhlfzsi%2Fnonebot_plugin_lazytea%2Fmain%2Fimages%2FREADME%2Fapp.png&name=1&owner=1&pattern=Plus&stargazers=1&theme=Auto" alt="LazyTea Mobile" width="640" height="320" />
</a>

_—— 将您的 Bot 管理面板，轻松放进兜里。 ——_

[![GitHub release](https://img.shields.io/github/v/release/hlfzsi/LazyTea-mobile.svg)](https://github.com/hlfzsi/LazyTea-mobile/releases)

</div>

**LazyTea Mobile** 是一个跨平台的移动应用程序 (Android/iOS)，它为 [LazyTea](https://github.com/hlfzsi/nonebot_plugin_lazytea) 后端服务提供了一个轻量、原生的移动管理界面。本项目使用 Go 语言和 [Fyne](https://fyne.io/) 框架构建。

# 📌 项目关系

为了更好地理解 LazyTea 生态，请注意以下几个项目的区别与联系：


| 项目                          | 主要职责                                       | 仓库地址                                                                   |
| :---------------------------- | :--------------------------------------------- | :------------------------------------------------------------------------- |
| 📱**LazyTea Mobile (本项目)** | **跨平台移动客户端**，用于连接并管理后端服务。 | [LazyTea-mobile](https://github.com/hlfzsi/LazyTea-mobile)                 |
| 💻**LazyTea Client**          | **独立桌面客户端**，功能更全面。               | [LazyTea-Client](https://github.com/hlfzsi/LazyTea-Client)                 |
| 🔌**nonebot_plugin_lazytea**  | **NoneBot2 插件**，作为 LazyTea 的后端服务端。 | [nonebot_plugin_lazytea](https://github.com/hlfzsi/nonebot_plugin_lazytea) |

简单来说，您需要在您的机器人上安装 `nonebot_plugin_lazytea` 作为服务端，然后使用本移动客户端连接到它，即可随时随地进行管理。

<br>

# ✨ 优势

* **原生移动体验**: 基于 Fyne 构建，在 Android 和 iOS 上提供流畅、一致的原生体验。
* **随时随地管理**: 摆脱地理位置限制，只要您的手机在身边，就能管理您的机器人。
* **轻量高效**: Go 语言保证了应用的性能和较低的资源占用。
* **核心功能**: 包含了插件管理、消息查看、Bot状态监控等核心功能。

# 🚀 快速入门

### **1. 前提条件**

1. **安装后端服务**: 确保您的 NoneBot2 项目中已正确安装并运行 `nonebot_plugin_lazytea` 插件。
2. **获取连接信息**: 启动 NoneBot2，从日志或配置文件中获取连接所需的 `IP`, `PORT` 和 `TOKEN`。请确保您的服务能在公网或局域网被您的手机访问到。

### **2. 安装应用**

#### 方式一：从发布版安装 (推荐)

1. 前往 [Releases](https://github.com/hlfzsi/LazyTea-mobile/releases) 页面。
2. 下载最新的 `.apk` 文件 (适用于 Android)。
3. 在您的 Android 设备上安装该 APK 文件。 (iOS 用户需要自行编译并通过 TestFlight 或自行签名安装 `.ipa` 文件)。

#### 方式二：从源码构建

如果您希望自行构建应用，请确保您已配置好 Go 和 Fyne 的开发环境。

**环境要求:**

* Go (1.17 或更高版本)
* Fyne CLI (`go install fyne.io/fyne/v2/cmd/fyne@latest`)
* Android SDK 和 NDK (用于构建 Android 应用)
* Xcode (在 macOS 上构建 iOS 应用)

**构建命令:**

```bash
# 克隆仓库
git clone https://github.com/hlfzsi/LazyTea-mobile.git
cd LazyTea-mobile

# 构建 Android APK
fyne package --os android -appID com.lazytea.mobile

# 构建 iOS 应用 (需要在 macOS 上执行)
fyne package --os ios -appID com.lazytea.mobile
```

构建完成后，生成的应用程序包将位于当前目录中。

# 🗺️ 项目蓝图

* [ ]  **UI/UX 优化**: 持续打磨移动端的用户交互和视觉体验。
* [ ]  **推送通知**: 探索实现关键事件（如Bot离线）的推送通知。
* [ ]  **性能优化**: 优化数据加载和渲染性能，提供更流畅的体验。
