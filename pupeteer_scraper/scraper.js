const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

async function scrape(url, cookiePath) {
    const browser = await puppeteer.launch({
        args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    const page = await browser.newPage();

    const cookiesString = fs.readFileSync(cookiePath);
    const cookies = JSON.parse(cookiesString);
    await page.setCookie(...cookies);

    await page.goto(url, {waitUntil: 'networkidle2'});

    const result = await page.evaluate(() => {
        let title = document.querySelector('meta[property="og:title"]')?.content;
        let image = document.querySelector('meta[property="og:image"]')?.content;
        let description = document.querySelector('meta[property="og:description"]')?.content;
        let author = document.querySelector('.tgme_widget_message_forwarded_from_name')?.innerText;
        let views = document.querySelector('.tgme_widget_message_views')?.innerText;
        let datetime = document.querySelector('time.datetime')?.getAttribute('datetime');
        let links = Array.from(document.querySelectorAll('a')).map(a => a.href);
        
        return { title, image, description, author, views, datetime, links };
    });

    console.log(JSON.stringify(result));

    await browser.close();
}

const [url, cookiePath] = process.argv.slice(2);
scrape(url, cookiePath).catch(console.error);
