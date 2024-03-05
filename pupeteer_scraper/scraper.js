const puppeteer = require('puppeteer');
const fs = require('fs');

async function scrape(url, cookiePath, isDebug) {
    const browser = await puppeteer.launch({
        args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    const page = await browser.newPage();

    if (cookiePath) {
        try {
            const cookiesString = fs.readFileSync(cookiePath, 'utf8'); 
            const cookies = JSON.parse(cookiesString);
            await page.setCookie(...cookies);
        } catch (error) {
            console.error(`Failed to load or parse cookies from ${cookiePath}:`, error);
            await browser.close();
            return;
        }
    }

    await page.goto(url, {waitUntil: 'networkidle2'});

    if (isDebug) {
        const content = await page.content(); 
        console.log(content);
    }

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

const url = process.argv[2];
const cookiePath = process.argv[3]; 
const isDebug = process.argv.includes('--debug');

scrape(url, cookiePath, isDebug).catch(console.error);
