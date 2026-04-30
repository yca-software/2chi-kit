import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "astro/config";

export default defineConfig({
	site: "https://example.com",
	integrations: [react(), sitemap()],
	output: "static",
	vite: {
		plugins: [tailwindcss()],
	},
});
