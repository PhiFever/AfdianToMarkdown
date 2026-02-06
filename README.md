![GitHub Repo stars](https://img.shields.io/github/stars/PhiFever/AfdianToMarkdown)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/PhiFever/AfdianToMarkdown/total)
## AfdianToMarkdown

爱发电(afdian.com)爬虫，用于下载爱发电作者的动态和作品集并保存为markdown文件（目前只能保存纯文本内容，不支持保存图片）。

**！！！该软件不能直接帮你免费爬取订阅后才能查看的内容！！！**

### 准备

使用浏览器插件[cookie master](https://chromewebstore.google.com/detail/cookie-master/jahkihogapggenanjnlfdcbgmldngnfl)导出爱发电cookie，如下图所示点击`copy`

![image](https://github.com/user-attachments/assets/d27b0f59-95c0-4080-97b9-d544d5424a33)

将复制到的json文本粘贴进与在RELEASE中下载的可执行文件同级（或git clone的项目根目录）的`cookies.json`即可。

![image](https://github.com/user-attachments/assets/3c9a4a26-fa94-4c38-a69d-359a536446b1)

**注意主站域名可能需要手动指定（默认为afdian.com)**

### 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--host` | 主站域名，如访问不通可自行更改 | `afdian.com` |
| `--dir` | 数据存储目录 | 程序所在目录下的 `data` 文件夹 |
| `--cookie` | cookies.json 文件路径 | 程序所在目录下的 `cookies.json` |
| `--disable_comment` | 为 true 时不下载评论 | `false` |

### 构建

如果你不需要对源码进行开发，请跳至下一节。

- go build

### 帮助
```
$ .\AfdianToMarkdown.exe -h
```

### 使用

本程序为命令行程序，需要在`cmd`,`powershell`或`bash`等shell中输入参数调用刚才构建的（或在release中下载的）可执行程序

- 在win11上，直接在已下载的可执行程序目录下右键，点击`open in terminal`或`在终端中打开`

  ![image](https://github.com/user-attachments/assets/94013988-2579-4fe2-a4b2-5245ad4501da)

- 在win10上，情况稍微复杂

  1. win+r，输入`powershell`

  2. 找到可执行文件所在的路径（如`D:\MyProject\Golang\AfdianToMarkdown\AfdianToMarkdown.exe`)

     输入`cd D:\MyProject\Golang\AfdianToMarkdown`

后续所有的指令都将在命令行中执行。在windows平台上进行调用时，有如下示例（即`AfdianToMarkdown.exe -h`）：

注：下文提到的`user_id`为作者主页url的最后一部分，如`https://afdian.com/a/user_id/`

#### 下载作者的所有动态

```shell
.\AfdianToMarkdown.exe motions --host="ifdian.net" -au "user_id" 
```

#### 下载作者所有的作品集

```shell
.\AfdianToMarkdown.exe albums -au "user_id" 
```

#### 更新所有已经下载的作者的动态和作品集
1. 需要对应的作者目录下具有`motions`目录
2. 不会覆盖已经下载的文件，所以也不会更新评论。可以通过删除文件来强制更新

```shell
.\AfdianToMarkdown.exe --host="ifdian.net" update
```

#### 下载任意作者的单个作品集

```shell
.\AfdianToMarkdown.exe album -u https://afdian.com/album/aaa
```

### 更新日志

#### v0.5.0

1. 添加了下载任意作者的单个作品集的功能
2. 添加在下载时不下载评论区的控制参数

#### v0.4.0
增加了对于含图片作品集的支持

#### v0.3.0
1. 修改默认域名为`afdian.com`
2. 将寻找`cookies.json`的逻辑修改为在程序目录下而非工作目录下
3. 修复了对域名`ifdian.net`解析不正确的问题

#### v0.2.2

缩短了等待时间，加快下载效率

#### v0.2.1

添加了对update指令的支持，修复了Refer中url不正确的问题

#### v0.2
由于主站(afdian.net)在7月15日被屏蔽，添加了对于手动更改临时域名(如afdian.com)的支持
