const fs = require('fs');
const PNG = require('pngjs').PNG;
const pixelmatch = require('pixelmatch');

/**
 * Compares two PNG screenshots and generates a diff image.
 * Requires: npm install pngjs pixelmatch
 */
function compareScreenshots(basePath, testPath, diffPath, threshold = 0.1) {
  try {
    const img1 = PNG.sync.read(fs.readFileSync(basePath));
    const img2 = PNG.sync.read(fs.readFileSync(testPath));

    const { width, height } = img1;
    
    if (width !== img2.width || height !== img2.height) {
      console.error('Image dimensions do not match.');
      return false;
    }

    const diff = new PNG({ width, height });

    const numDiffPixels = pixelmatch(
      img1.data, 
      img2.data, 
      diff.data, 
      width, 
      height, 
      { threshold }
    );

    fs.writeFileSync(diffPath, PNG.sync.write(diff));

    const diffPercentage = (numDiffPixels / (width * height)) * 100;
    console.log(`Matched with ${diffPercentage.toFixed(2)}% difference.`);
    
    return numDiffPixels === 0;

  } catch (err) {
    console.error('Error during comparison:', err.message);
    return false;
  }
}

// Example usage
// compareScreenshots('baseline.png', 'current.png', 'diff.png');
