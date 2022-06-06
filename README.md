NBlog
======

一个以 Notion 作为后端的简易博客。

如何使用
------

1. 新建一个页面，并在页面中创建一个 `Database - Full page`，并且按下面的格式创建列。
2. 根据[指引](https://developers.notion.com/docs/getting-started)来创建可以正常使用的 API Key。
3. 根据[文档](https://developers.notion.com/docs/working-with-databases#adding-pages-to-a-database)来获取 Database ID。
4. 将 API Key 和 Database ID 填入 `config.json`，或者以 `apikey` 和 `database` 作为环境变量名称运行程序。

Notion 数据库格式
------
``` json
[
    {"name":"Title", "type":"Title"},
    {"name":"Published", "type":"Checkbox"},
    {"name":"Content", "type":"Text"},
    {"name":"Time", "type":"Date"}
]
```