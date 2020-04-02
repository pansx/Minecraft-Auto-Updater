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

2. 如果是初次制作更新包，附加参数`--init`来获得

3. 运行更新器，附加参数`--pack`

4. 删除服务器上的`file_info.json`

5. 上传`./package`文件夹中的`download`文件夹和`file_info.json`

6. 将`package/update_info.json`编辑后上传：

   说明：

   - `"version":int`游戏版本，**每次更新后应将版本号加一**，这个任务会由更新器自动完成

   - `"ignore_list":[]string`普通更新过程中被忽略的文件的**前缀**，按需编辑，例如：
     
     - `"game/hmcl.json"`意味着忽略`"game/hmcl.json"`文件
     
     - `"game/screenshots"`意味着忽略`"game/screenshots"`文件夹下所有文件
     
       **请注意！`ignore_list`中的分隔符一定要是`/`**
     
   - `"package_ignore_list":[]string`打包过程中要忽略的文件的前缀，如上

     你可能要问，这两个list有什么区别？`ignore_list`当中的文件会被打包，普通更新时不被更新，但是修复模式下会无视它，仍然参与更新；`package_ignore_list`当中的文件不会被打包，任何时候都不会参与更新

     举几个例子来做一个说明：

     比如当前`game`目录下有a,b,c,d四个文件，其中a不在两个列表里，b在`ignore_list`里，c在`package_ignore_list`，d在`ignore_list`和`package_ignore_list`里

     进行普通更新时，a会被更新到和服务器上一样的版本，b,d则不会，而c会被删除

     在修复模式下，a,b会被更新到和服务器上一样的版本，d则不会，而c会被删除

     一般来说，mod文件、mod配置文件属于a一类的文件，需要每次都更新；用户配置文件则属于b类文件；你游戏文件夹里不想给用户的私人文件属于c类，但你把它当作d类也可以；日志文件和截图文件、存档文件每个人都不一样且不影响游戏，而启动器配置文件因为包含了你的登录信息，我们不想让它上传，但也不想覆盖掉用户设置，这些属于d类文件
   
7. 如果要让更新器自动启动启动器，请用纯java启动器（如HMCL），并将其名称改为`Launcher.jar`（注意大小写）

8. 如果要修改`update_info.json`，请修改服务器端的，否则任何在本地作的修改都不会生效

`update_info.json`示例：

`{"version": 1, "ignore_list": ["game/hmcl.json","game/.minecraft/screenshots","game/.minecraft/options.txt"], "package_ignore_list": ["game/hmcl.json","game/screenshots"]}`

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