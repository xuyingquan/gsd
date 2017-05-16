# Global Service Dispatcher 5
## 简介
全局负载均衡器V5，支持以下特性：

* DNS协议，基于算法返回需要负载均衡的入口IP A记录 CNAME记录
* 支持域名前缀模糊匹配，匹配从精确到模糊（.开头域名为模糊匹配，如.shatacloud.com匹配*.shatacloud.com）
* 支持基于IP来源规则返回 A记录 CNAME记录
* 支持按权重打散返回ip
* 支持从SDN API获得IP后返回结果

## 构建
进入项目根目录

执行`./build.sh`

构建成功后，可看到打包文件`gsd5.tar.gz`

拷贝到安装目录，解压后执行`bin/gsd version`可查看版本号

注：如果出现长时间卡在"get denpendency"环节，说明GFW在干扰编译，这个时候可以用以下方法解决

在gsd5同级目录下执行(即用GolangDeps下的目录覆盖gsd5下的同级目录)

`git clone git@git.shatacloud.com:rd/GolangDeps.git`

`cp -R GolangDeps/src gsd5`

然后按上述步骤构建即可

## 执行
执行样例：

前台执行

`gsd run -l :53 -c /opt/gsd5/etc`

后台执行

`gsd start -l :53 -c /opt/gsd5/etc`

* -l: 绑定的IP和端口，:53代表绑定所有IP的53端口
* -c: 配置文件所在跟目录
* -v: 显示版本号
* -h: 显示说明

后台执行时，日志默认输出在logs/gsd.bin.out文件，如需修改日志输出路径，在环境变量中指定GSD_OUT环境变量即可，例如

`export GSD_OUT=/var/log/gsd5/gsd.out`

### 配置文件说明

配置文件格式etc/README.md

### 日志级别：
日志输出在标准输出下，输出的日志级别在环境变量中指明

如下样例表示以文本格式输出所有INFO级别日志，日志的级别有ERR、INF、DBG，TRC等，详细设置见[logxi项目页](https://www.github.com/mgutz/logxi)

`export LOGXI=*=INF`

`export LOGXI_FORMAT=text`

启动脚本默认级别为INF

