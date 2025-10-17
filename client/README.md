# 启动项目

## 安装环境

```shell
    npm install
```

# 启动

```shell
    npm run dev
```

# 打包

```shell
    npm run build
```
+ 注意，打包时 `index.js` 首行请用：`const API_BASE_URL = '/api'`
+ 本地测试时， `indxex.js` 首行请用：`const API_BASE_URL = 'http://localhost:8123/api'`
+ 部署时记得将 Nginx 配置好