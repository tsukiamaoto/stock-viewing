const https = require('https');

https.get('https://www.cmoney.tw/forum/stock/1711', { 
    headers: { 'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36' } 
}, (res) => {
    let raw = '';
    res.on('data', c => raw += c);
    res.on('end', () => {
        // Look for the NUXT state script block
        const st = raw.indexOf('window.__NUXT__=');
        if (st === -1) {
            console.log('No NUXT found');
            return;
        }
        let end = raw.indexOf('</script>', st);
        let script = raw.slice(st, end);
        console.log('Script size:', script.length);
        
        // We know it is window.__NUXT__=(function(a,b,c,...){ return {...}; }(arg1,arg2,...));
        // Find the args list
        const argsMatch = script.match(/\}\((.*)\)\);$/);
        if (argsMatch) {
            console.log('Extracted args!', argsMatch[1].slice(0, 50));
        }

        // Just use a poor man's search to see if we can find posts
        // For instance, looking for names, titles, views
        let postStr = raw.match(/\"Title\":\"(.*?)\"/g);
        if(postStr) {
            console.log("Found posts with pure regex:", postStr.length);
        } else {
            console.log("No Title found. Let's look for known fields...");
            let field = raw.match(/"?content"?:/i);
            if(field) {
                console.log("Found field content");
            }
        }
    });
});
