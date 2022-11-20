### alerter     

根据告警规则配置信息，处理指标数据判断告警，告警分发。

.
├── cmd                   ## 可独立执行程序

│   ├── ...

├── configs               ## 配置文件

├── docs

├── internal              ## 业务目录

│   ├── api               ## api 接口

│   ├── code              ## 错误码定义

│   ├── pkg               ## 内部使用的 package

│   ├── repository        ## 实体

│   ├── router            ## 路由

│   ├── services          ## 服务

│   └── websocket         ## websocket 接口

├── logs                  ## 日志目录

├── pkg                   ## 可供外部使用的 package

│   ├── ...

└── scripts               ## 脚本程序
