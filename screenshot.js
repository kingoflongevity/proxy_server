const puppeteer = require('puppeteer');

async function screenshot() {
  const browser = await puppeteer.launch({
    headless: true,
    defaultViewport: {
      width: 1920,
      height: 1080
    },
    args: ['--no-sandbox', '--disable-setuid-sandbox']
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
    try {
      await page.goto(`http://localhost:3001/${pageInfo.name}`);
      await page.waitForFunction(() => true, { timeout: 3000 }); // 等待页面加载
      await page.screenshot({
        path: `screenshots/${pageInfo.name}.png`,
        fullPage: true
      });
      console.log(`已截图: ${pageInfo.title}`);
    } catch (error) {
      console.error(`截图 ${pageInfo.title} 失败: ${error.message}`);
    }
  }
  
  await browser.close();
  console.log('所有页面截图完成');
}

screenshot();
