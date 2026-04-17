## Quick Start

### English

1. **Download**
    - Get the latest package from [Releases](https://github.com/hj6hki123/ssm-gui/releases) and extract it.
    - If you already have the original `ssm` project, you can place `ssm-gui.exe` in the same folder.

2. **Start the program**
    - Double-click `ssm-gui.exe`, or run:
      ```bash
      ./ssm-gui.exe
      ```
    - The UI should open automatically at `http://127.0.0.1:8765`.
    - If it does not open, visit the address manually in your browser.

3. **Prepare your phone**
    - Connect your phone to your PC with a USB cable.
    - Enable **USB debugging / ADB debugging** on the phone.

4. **Prepare Game Resources**
    - Ensure the game is installed and you have downloaded all in-game content.
    - Copy and extract the game resource pack to your computer.
    - Also copy the device data directory:
      **Using ADB to copy:**
      - BanG Dream (JP) example:
         ```bash
         adb pull /sdcard/Android/data/jp.co.craftegg.band/files/data/ ./gamedata
         ```
         *This command will copy the game data directory to the `/gamedata` folder in your current computer location*
      - Generic path format:
         `/sdcard/Android/data/{game_package_name}/files/data/`
    - **Use SSM to decode game chart resources:**
      - Run `ssm-gui.exe` and access the web UI.
      - Click **Extract Assets**.
      - Enter the path to the game data copied in the previous step (e.g., `D:/gamedata`) and click **Extract**.
      - The chart files will be automatically extracted to the `/assets` folder in the project.

5. **Set up device in GUI**
    - Open the **Settings** page.
    - Add your device (serial number can be auto-detected or selected from dropdown).
    - Choose connection type: **HID** or **ADB**.

6. **Load song and start playback**
    - In the main flow, go through: **Song Setup -> Play Control -> Start**.
    - When the first note reaches the judgement line, press **Start** (or keyboard **Enter** / **Space**).
    - If timing is early/late, adjust **Offset/Delay** and retry.

> Legacy command-line usage is still supported. You can append original CLI parameters as before.
> See [kvarenzn's Usage Guide](https://github.com/kvarenzn/ssm/blob/main/docs/USAGE.md).

---
### 简体中文

1. **下载与解压**

   * 从 [Releases](https://github.com/hj6hki123/ssm-gui/releases) 下载最新版本并解压。
   * 如果你已经有原版 `ssm` 项目，直接把 `ssm-gui.exe` 放到同一文件夹即可。

2. **启动程序**

   * 直接双击 `ssm-gui.exe`，或用终端运行：

     ```bash
     ./ssm-gui.exe
     ```
   * 程序会尝试自动打开浏览器到 `http://127.0.0.1:8765`。
   * 如果没有自动打开，请手动输入网址。

3. **连接并准备手机或模拟器**

   * 手机请用 USB 线连接电脑；如果使用模拟器，请先确认 ADB 已启用。
   * 在手机上开启 **USB 调试 / ADB 调试**。
   * 可使用以下命令确认设备是否已连接：

     ```bash
     adb devices
     ```

4. **准备游戏资源**
   * 确保已经安装游戏并且进入游戏下载内容
   * 将游戏资源包复制到电脑并解压
   * 同时把手机中的数据目录复制到电脑：
    **使用ADB复制:**
     * BanG Dream (JP)示例：

       ```bash
       adb pull /sdcard/Android/data/jp.co.craftegg.band/files/data/ ./gamedata
       ```
       *此命令会将游戏数据目录复制到当前电脑位置中的`/gamedata`文件夹中*
     * 通用路径：
       `/sdcard/Android/data/{游戏包名}/files/data/`
    * **使用SSM解码游戏谱面资源**
      * 运行`ssm-gui.exe`, 进入web ui
      * 点击<解包素材(Extract Assets)>
      * 输入上一步复制到电脑的游戏数据的位置目录，如`D:/gamedata`, 点击<提取(Extract)>
      * 谱面文件会自动提取到项目文件夹的`/assets`里

5. **在 GUI 中设置设备**

   * 进入 **Settings** 页面添加设备。
   * 序列号可以自动检测，或从下拉菜单选择。
   * 连接方式选择 **HID** 或 **ADB**。

6. **选歌并开始**

   * 按流程操作：**Song Setup -> Play Control -> Start**。
   * 当第一个音符接近判定线时，按 **Start**（或键盘 **Enter** / **Space**）。
   * 如果时机偏早或偏晚，可以调整 **Offset/Delay** 后重试。

> 仍可使用传统命令行参数方式启动。
> 详细参数请参考 [kvarenzn 的使用指南](https://github.com/kvarenzn/ssm/blob/main/docs/USAGE.md)。
