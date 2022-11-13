## 🎡 <font color="green">介绍</font>

> all eyes是一个拥有自定义监控能力，无需Agent的实时监控系统。网站监测，PING连通性，端口可用性，数据库，操作系统，中间件，API监控，阈值告警，告警通知(邮件微信钉钉飞书)。 


## 🥐 模块  

- **[manager](https://github.com/dromara/hertzbeat/tree/master/manager)** 提供监控管理,系统管理基础服务
> 提供对监控的管理，监控应用配置的管理，系统用户租户后台管理等。
- **[collector](https://github.com/dromara/hertzbeat/tree/master/collector)** 提供监控数据采集服务
> 使用通用协议远程采集获取对端指标数据。
- **[warehouse](https://github.com/dromara/hertzbeat/tree/master/warehouse)** 提供监控数据仓储服务
> 采集指标结果数据管理，数据落盘，查询，计算统计。
- **[alerter](https://github.com/dromara/hertzbeat/tree/master/alerter)** 提供告警服务
> 告警计算触发，监控状态联动，告警配置，告警通知。
- **[web-app](https://github.com/dromara/hertzbeat/tree/master/web-app)** 提供可视化控制台页面
> 监控告警系统可视化控制台前端(angular+ts+zorro)  

![hertzBeat](https://tancloud.gd2.qingstor.com/img/docs/hertzbeat-stru.svg)   

