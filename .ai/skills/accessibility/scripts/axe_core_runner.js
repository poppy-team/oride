const puppeteer = require('puppeteer');
const { AxePuppeteer } = require('@axe-core/puppeteer');
const fs = require('fs');

async function runAxeTest(url, outputPath) {
  console.log(`Starting accessibility analysis for: ${url}`);
  const browser = await puppeteer.launch({ headless: "new" });
  
  try {
    const page = await browser.newPage();
    await page.setBypassCSP(true);
    await page.goto(url, { waitUntil: 'networkidle2' });

    console.log('Running Axe Core...');
    const results = await new AxePuppeteer(page)
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .analyze();

    const report = {
      url: results.url,
      timestamp: results.timestamp,
      violations: results.violations.map(v => ({
        id: v.id,
        impact: v.impact,
        description: v.description,
        help: v.help,
        helpUrl: v.helpUrl,
        nodes: v.nodes.length
      }))
    };

    fs.writeFileSync(outputPath, JSON.stringify(report, null, 2));
    console.log(`Report saved to ${outputPath}`);
    console.log(`Found ${report.violations.length} violation rules.`);

  } catch (error) {
    console.error('Error running axe-core:', error);
  } finally {
    await browser.close();
  }
}

// Example usage:
// node axe_core_runner.js https://example.com report.json
const args = process.argv.slice(2);
if (args.length >= 2) {
    runAxeTest(args[0], args[1]);
} else {
    console.log("Usage: node axe_core_runner.js <url> <output_file>");
}
