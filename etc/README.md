# 配置说明
## 概述

GSD5的配置文件主要采用YAML格式描述，由若干组配置文件构成，具体构成如下：

### etc/gsd.yaml
gsd全局配置，样例如下：

    redis: true
    redisAddr: "localhost:6379"
    redisPass: ""

具体说明：

1. `redis`: 是否启用redis后端作为配置点，如果为true，则其它配置文件除ipdb外不再生效，只启用redis中的配置

2. `redisAddr`: redis连接地址

3. `redisPass`: redis连接密码，无则填写""

### etc/ipdb

ipdb文件为ip与区域、运营商关系对应描述文件，该文件为csv格式，具体字段描述如下：

> 起始ip,结束ip,所属区域,所属运营商

样例如下：

> 16785408,16793599,中国/广东省/广州市/广州市其他县,中国电信

该文件的ip必须按从小到大排列，地域描述按“国家/省份/地级市/县”格式描述

### hosts/*.yaml

hosts目录下的配置文件描述所有的域名，文件名可以为任意名称，但必须以yaml为后缀，GSD会加载该目录下所有后缀为yaml的文件，
应避免在不同文件中重复定义同一域名，否则GSD可能出现未知状况。

hosts配置样例如下：

    "test.shatacloud.com":
      record: A
      ttl: 60
      loadbalance: policy
      policy: 性能优先
      ipkey: pub
      max: 4
      label:
        channel: test
      
  
具体说明：

1. `"test.shatacloud.com"`: 配置的域名，以“.”开头时前缀模糊匹配，例如：
`.shatacloud.com`代表所有以.shatacloud.com结尾的域名
  
    当同时有"test.shatacloud.com"和".shatacloud.com"配置时，以更精确的配置优先匹配。

2. `record: A` : 返回的记录类型，共有两种
 
    * A
    * CNAME
    
3. `ttl: 60` : ttl时间， 单位为秒

4. `loadbalance: policy` : 负载均衡策略，目前支持5种策略，分别为：
    
    * policy 根据policy（按区域调度策略）配置决策(policies目录中的配置文件)    
        
    * sdn_api 根据sdn后端的返回值决定
        
        当配置为此项时，会有扩展配置  
            regex: (.+)\\.sdn
            apiurl: "http://127.0.0.1:10001/push?ip=$remoteIp&app=sdn&instance=$1"
        含义为，按正则表达式生成访问SDN后端的url
        
    * random 返回target中配置项作为结果，返回顺序随机
    * 其它 即此项设成任意其它值时，会返回target中配置项作为结果，返回顺序与配置顺序相同
        
5. `ipkey: pub` :record为A时生效，即A记录中的IP取值Pool配置中的哪一项，具体见Pools配置描述

6. `max: 4` :record为A时生效，返回的A记录中的IP最多可以返回多少条，不填时代表不限制

7. `target:` : loadbalance策略为consistent_hash、 random、all或其它时生效，其中：
        `pool:` : 候选入口地址的pool名称，名称需与pools中的配置一致
        `weight:` : 候选地址pool出现在第一位的权重，当loadbalance策略为random是生效

8. `label:` : 用于标记改设置的其它标签，可用于统计（范例中为channel，可为任意值）
 
### pools/*.yaml

pools目录下的配置文件描述所有的pool，pool代表一组入口的集合，文件名可以为任意名称，但必须以yaml为后缀，
GSD会加载该目录下所有后缀为yaml的文件，应避免在不同文件中重复定义同一个pool名，否则GSD可能出现未知状况。

pools配置样例如下：

    嘉兴电信:
      - name: jx-l1-1
        ip:
          pub: 115.231.73.130
          priv: 10.0.1.1
        weight: 1
      - name: jx-l1-2
        ip:
          pub: 115.231.73.131
    杭州电信:
      - name: hz-l1-1
        ip:
          pub: 115.236.57.99
      - name: hz-l1-2
        ip:
          pub: 115.236.57.98
    华东区电信:
        - cname: hd.tl.shatacloud.com
          
说明如下：

1. `嘉兴电信:` : Pool名称，可以为任意值

2. `- name: jx-l1-1` : 入口点名称，可以为任意值，也可为空

3. `ip:` : 入口的ip地址，可以有多个，分别以不同的名称(名称可任意定义)做标识，在hosts中通过ipkey选择

4. `cname:` : 入口的cname地址，当hosts配置的records为cname时，将返回该值

5. `weight:` : 权重，即当出现需要打散返回时，该入口出现在第一行的几率，默认值为1 
 
### policies/*.yaml

policies目录下的配置文件描述所有按区域调度的策略配置，文件名可以为任意名称，但必须以yaml为后缀，
GSD会加载该目录下所有后缀为yaml的文件，应避免在不同文件中重复定义同一个policy名称，否则GSD可能出现未知状况。

配置样例如下:

    性能优先:
      中国/上海市@中国电信:
      - pool: 嘉兴电信
        weight: 2
        priority: 1
        disable: false
      - pool: 杭州电信
        weight: 1
      default@中国电信:
      - pool: 嘉兴电信
      default@default:
      - pool: 杭州电信
          
说明如下：

1. `性能优先` : policy名称，在hosts中的policy配置可指定此名称选取策略

2. `中国/上海市@中国电信:` ： 区域和运营商， 区域间用/分隔，区域与运营商间用@分隔，区域采用后缀匹配原则，区域和运营商有remote 
    即：中国/上海市/浦东新区 将匹配 中国/上海市 ， 除非有更精确的配置
    
3. `pool: 嘉兴电信` : 选取的入口pool名称

4. `weight: 2` : 选取权重，当hosts返回A记录时生效，它决定了一个pool在多次请求中出现在ip列表首位的概率。

5. `priority: 1`: 优先级，当有可用的高优先级pool时，不选择低优先级pool，优先级的值越小代表优先级越高

6. `disable: false`: pool是否停用，默认值为false

7. `default@中国电信` : 默认区域为default, 当ip匹配不上任何区域是匹配该配置。

8. `default@default` : 默认运营商为default, 当ip匹配不上任何区域和运营商时匹配改配置。

### zones/*.yaml

zones目录下配置所有的ns记录配置，文件名可以为任意名称，但必须以yaml为后缀，
GSD会加载该目录下所有后缀为yaml的文件，应避免在不同文件中重复定义同一个zone名称，否则GSD可能出现未知状况。

配置样例如下：

    "okey.com":
      origin: okey.com
      soa: "aaa.com 500 IN SOA ns1.aaa.com. root.aaa.com.  42 3600 3600 360000 60"
      ns:
        - name: ns1.okey.com
          ip: 1.1.1.1
        - name: ns2.okey.com
          ip: 1.1.1.2

说明如下：

1. "okey.com": 配置的域名

2. "origin": ns记录中的origin

3. "soa": ns中的soa
 
4. "ns": 所有的ns记录值