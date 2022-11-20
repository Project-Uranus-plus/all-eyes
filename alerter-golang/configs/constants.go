package configs

const (
	// MinGoVersion 最小 Go 版本
	MinGoVersion = 1.16

	// ProjectVersion 项目版本
	ProjectVersion = "v1.0.0"

	// ProjectName 项目名称
	ProjectName = "ALL-EYES-ALERTER"

	// ProjectDomain 项目域名
	ProjectDomain = "http://127.0.0.1"

	// ProjectPort 项目端口
	ProjectPort = ":9999"

	// ProjectAccessLogFile 项目访问日志存放文件
	ProjectAccessLogFile = "./logs/" + ProjectName + "-access.log"

	// ProjectCronLogFile 项目后台任务日志存放文件
	ProjectCronLogFile = "./logs/" + ProjectName + "-cron.log"

	// ProjectInstallMark 项目安装完成标识
	ProjectInstallMark = "INSTALL.lock"

	// ZhCN 简体中文 - 中国
	ZhCN = "zh-cn"

	// EnUS 英文 - 美国
	EnUS = "en-us"

	// MaxRequestsPerSecond 每秒最大请求量
	MaxRequestsPerSecond = 10000

)
