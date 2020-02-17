# Minecraft自动更新器说明（MAU）

## 用户

### 初次使用

1. 新建一个**空文件夹**用于存放游戏
2. 将更新器放入其中
3. 双击运行或命令行运行（不附加参数）
4. 等待程序自动退出后，进入`./game`文件夹运行游戏即可
5. 如果你有文件放在./game中被误删的（**并不建议将任何个人文件放在游戏目录中**），可以去./rubbish文件夹中寻找，否则可以将./rubbish删除以节省磁盘空间

### 更新游戏

同上3~5

### 修复游戏文件

若游戏出错则**用命令行运行更新器**并附加`--repair`参数，**或删除`update_info.json`后运行更新器**

## 开发维护人员

### 制作更新包

1. 在`./game`文件夹中添加你的游戏文件

2. 清除`./package`下的所有内容（如果有的话）

3. 运行更新器，附加参数`--pack`

4. 删除服务器上的`download`文件夹和`file_info.json`

5. 上传./package文件夹中的`download`文件夹和`file_info.json`

6. 将`update_info.json`下载编辑后上传：

   说明：

   - `"version":int`游戏版本，**每次更新后应将版本号加一**
   - `"updater_version":int`更新器版本，不用动它
   - `"resource_url":string`资源文件地址，不用动它
   - `"ignore_list":[]string`普通更新过程中被忽略的文件的**前缀**，按需编辑，例如：
     
     - `"game/hmcl.json"`意味着忽略`"game/hmcl.json"`文件
     
     - `"game/screenshots"`意味着忽略`"game/screenshots"`文件夹下所有文件
     
       **请注意！`ignore_list`中的分隔符一定要是`/`**

7. 完成更新

`update_info.json`示例：

`{"version": 1, "updater_version": 1, "resource_url": "https://www.yoursite.com/", "ignore_list": ["game/hmcl.json","game/screenshots"]}`

### 搭建你自己的更新服务

首先，搭建一个文件服务器，比如可以用阿里云oss，其目录结构应为：

- `https://www.yoursite.com/`
  - `download/`----游戏文件，由更新器自动生成
    - `(hash).zip`----游戏文件，由更新器自动生成
  - `update_info.json`----按照前文方法编辑
  - `file_list.json`----游戏文件hash列表，由更新器自动生成

然后，克隆这个git仓库，**修改`AutoUpdater.go`中的常量`resourceURL`为你的文件服务器地址**，保存，例如：

`const resourceURL = "https://minecraft-updater.oss-cn-shanghai.aliyuncs.com/"`**末尾有斜杠**

最后，确保你安装了go的编译器，（我是在1.13.5版本下编译通过的），运行build.bat或自行编译即可获得更新器，无需安装额外的包