# Icons

This directory contains PWA icons in the following sizes:
- icon-72x72.png
- icon-96x96.png
- icon-128x128.png
- icon-144x144.png
- icon-152x152.png
- icon-192x192.png
- icon-384x384.png
- icon-512x512.png

## Regenerating Icons

To regenerate these icons from the SVG favicon, run:

```bash
npm run generate-icons
# or
pnpm generate-icons
```

This will convert `public/favicon.svg` into all required PNG sizes.

Alternatively, you can use online tools:
- https://realfavicongenerator.net/
- https://www.pwabuilder.com/imageGenerator
- https://favicon.io/
