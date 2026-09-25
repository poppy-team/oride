const fs = require('fs');
const readline = require('readline');

/**
 * Parses a simple CSS file and extracts CSS variables (custom properties)
 * into a structured JSON Design Token format.
 */
async function extractTokens(cssFilePath, outputPath) {
  const fileStream = fs.createReadStream(cssFilePath);
  const rl = readline.createInterface({
    input: fileStream,
    crlfDelay: Infinity
  });

  const tokens = {
    colors: {},
    spacing: {},
    typography: {}
  };

  const cssVarRegex = /--([a-zA-Z0-9-]+):\s*([^;]+);/;

  for await (const line of rl) {
    const match = line.match(cssVarRegex);
    if (match) {
      const name = match[1];
      const value = match[2].trim();

      if (name.includes('color')) {
        tokens.colors[name] = { value, type: "color" };
      } else if (name.includes('space') || name.includes('padding') || name.includes('margin')) {
        tokens.spacing[name] = { value, type: "dimension" };
      } else if (name.includes('font') || name.includes('text')) {
        tokens.typography[name] = { value, type: "font" };
      }
    }
  }

  fs.writeFileSync(outputPath, JSON.stringify(tokens, null, 2));
  console.log(`Design tokens extracted to ${outputPath}`);
}

// Example execution
const sampleCSS = `
:root {
  --color-primary: #007bff;
  --color-secondary: #6c757d;
  --spacing-small: 8px;
  --spacing-medium: 16px;
  --font-base: 16px;
}
`;
fs.writeFileSync('temp_style.css', sampleCSS);
extractTokens('temp_style.css', 'design_tokens.json');
