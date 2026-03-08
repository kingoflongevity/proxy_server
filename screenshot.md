# 前端页面截图脚本

## 页面列表
- 仪表盘 (Dashboard)
- 订阅管理 (Subscriptions)
- 节点列表 (Nodes)
- 集群管理 (Cluster)
- 规则配置 (Rules)
- 系统设置 (Settings)
- 流量日志 (Logs)

## 截图命令

使用Chrome浏览器自动截图：
```bash
# 安装puppeteer
npm install -g puppeteer

# 运行截图脚本
node screenshot.js
```

## 截图脚本 (screenshot.js)

```javascript
const puppeteer = require('puppeteer');

async function screenshot() {
  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: {
      width: 1920,
      height: 1080
    }
  });
  
  const pages = [
    { name: 'dashboard', title: '仪表盘' },
    { name: 'subscriptions', title: '订阅管理' },
    { name: 'nodes', title: '节点列表' },
    { name: 'cluster', title: '集群管理' },
    { name: 'rules', title: '规则配置' },
    { name: 'settings', title: '系统设置' },
    { name: 'logs', title: '流量日志' }
  ];
  
  const page = await browser.newPage();
  
  for (const pageInfo of pages) {
    await page.goto(`http://localhost:3000/${pageInfo.name}`);
    await page.waitForTimeout(2000); // 等待页面加载
    await page.screenshot({
      path: `screenshots/${pageInfo.name}.png`,
      fullPage: true
    });
    console.log(`已截图: ${pageInfo.title}`);
  }
  
  await browser.close();
  console.log('所有页面截图完成');
}

screenshot();
```
