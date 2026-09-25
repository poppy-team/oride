# Example: SBOM Generation

A Software Bill of Materials (SBOM) provides visibility into the software supply chain. Here is how to generate them.

## Generating a CycloneDX SBOM for Node.js

You can use the `@cyclonedx/cyclonedx-npm` tool to generate an SBOM from a Node.js project.

```bash
# Install the tool globally
npm install -g @cyclonedx/cyclonedx-npm

# Run it in your project directory
cyclonedx-npm --output-file sbom.json
```

**Snippet of resulting `sbom.json`:**
```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "version": 1,
  "components": [
    {
      "type": "library",
      "name": "express",
      "version": "4.18.2",
      "description": "Fast, unopinionated, minimalist web framework",
      "hashes": [
        {
          "alg": "SHA-512",
          "content": "5/1uAHhN8nNq2tC...=="
        }
      ],
      "licenses": [
        {
          "license": {
            "id": "MIT"
          }
        }
      ],
      "purl": "pkg:npm/express@4.18.2"
    }
    // ... hundreds of other dependencies ...
  ]
}
```

## Generating a CycloneDX SBOM for Python

For Python projects, use `cyclonedx-bom`.

```bash
# Install the tool
pip install cyclonedx-bom

# Generate based on requirements.txt
cyclonedx-py requirements requirements.txt -o sbom.json
```

The resulting `sbom.json` can be uploaded to vulnerability management platforms (like Dependency-Track) for continuous monitoring against new CVE databases.
